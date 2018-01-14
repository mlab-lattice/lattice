/**
Package language provides json template language.

Main features:
 - Read/Eval json templates from git repositories
 - Single objection evaluation
 - Including other relative/external templates
 - Passing parameters
 - Variable declaration
 - String interpolation

Features to be added:
 - if statements

Interface:
 - EvalFromURL(url, parameters, options)
 - Eval(interface{}, parameters, options)


TODO:


 * Keep line numbers, maybe build a syntax tree where we have the preserve the property path and line number

 * Git urls: Which form do we want to allow?
  - If we allow ssh:// , what does that imply?
  - I am thinking allow file:// and git://
  - Whatever format we use, it should be allowed by the git/go-git as is

* Allow escaping in string interpolations

*/
package language

import (
	"github.com/mlab-lattice/system/pkg/util/git"
)

// Options evaluation options
type Options struct {
	GitOptions *git.Options // git options to be passed for git resolver
}

// TemplateEngine the main class to be used for parsing/evaluating templates
type TemplateEngine struct {
	operatorEvaluators map[string]OperatorEvaluator // internal operator evaluator registry
	operatorEvalOrder  []string                     // orders of operator evaluation within a template.
	// All keys operatorEvaluators are required to be here in the desired order.
}

// NewEngine constructs new engine object
func NewEngine() *TemplateEngine {

	engine := &TemplateEngine{
		operatorEvaluators: map[string]OperatorEvaluator{
			"$parameters": &ParametersEvaluator{},
			"$variables":  &VariablesEvaluator{},
			"$include":    &IncludeEvaluator{},
		},
		operatorEvalOrder: []string{
			"$parameters",
			"$variables",
			"$include",
		},
	}

	return engine
}

// EvalFromURL evaluates the template from the specified url with the specified parameters and options
func (engine *TemplateEngine) EvalFromURL(url string, parameters map[string]interface{}, options *Options) (map[string]interface{}, error) {
	env := newEnvironment(engine, options)
	result, err := engine.include(url, parameters, env)
	if err != nil {
		return nil, err
	}

	return result.(map[string]interface{}), nil
}

// Eval evaluates a single object
func (engine *TemplateEngine) Eval(o interface{}, parameters map[string]interface{},
	options *Options) (interface{}, error) {

	// create env and push parameters to the stack
	env := newEnvironment(engine, options)
	env.push("", parameters, make(map[string]interface{}))

	// call eval with env
	result, err := engine.eval(o, env)

	// pop
	env.pop()

	return result, err
}

// eval evaluates the specified object
func (engine *TemplateEngine) eval(o interface{}, env *environment) (interface{}, error) {

	if valMap, isMap := o.(map[string]interface{}); isMap { // Maps
		return engine.evalMap(valMap, env)

	} else if valArr, isArray := o.([]interface{}); isArray { // Arrays
		return engine.evalArray(valArr, env)

	} else if stringVal, isString := o.(string); isString { // Strings
		return engine.evalString(stringVal, env)

	} else { // Default, just return the value as is
		return o, nil
	}

}

// include includes and evaluates the template file specified in the url
func (engine *TemplateEngine) include(url string, parameters map[string]interface{}, env *environment) (interface{}, error) {
	// resolve url
	resource, err := resolveUrl(url, env)

	if err != nil {
		return nil, err
	}

	// init parameters if not set
	if parameters == nil {
		parameters = make(map[string]interface{})
	}
	// init variables
	variables := make(map[string]interface{})

	// push !
	env.push(resource.baseUrl, parameters, variables)

	// defer a pop to ensure that the stack is popped  before
	defer env.pop()

	// evaluate data of the template
	return engine.eval(resource.data, env)
}

// evalMap evaluates a map of objects
func (engine *TemplateEngine) evalMap(m map[string]interface{}, env *environment) (interface{}, error) {

	// init result
	result := make(map[string]interface{})
	// first, evaluate operators based on their priorities

	for _, operator := range engine.operatorEvalOrder {
		if operand, operatorExists := m[operator]; operatorExists {
			evaluator := engine.operatorEvaluators[operator]
			evalResult, err := evaluator.eval(operand, env)

			if err != nil {
				return nil, err
			} else if evalResult != nil {
				resultMap := evalResult.(map[string]interface{})
				// stuff map with the val result
				for k, v := range resultMap {
					result[k] = v
				}
			}
		}

	}
	// eval the rest of the map
	for k, v := range m {

		// skip operators since we have evaluated them already
		if _, isOperator := engine.operatorEvaluators[k]; isOperator {
			continue
		}

		var err error

		result[k], err = engine.eval(v, env)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

// evalArray evaluates an array of objects
func (engine *TemplateEngine) evalArray(arr []interface{}, env *environment) ([]interface{}, error) {
	result := make([]interface{}, len(arr))
	for i, v := range arr {
		var err error
		result[i], err = engine.eval(v, env)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

// evaluates a string
func (engine *TemplateEngine) evalString(s string, env *environment) (interface{}, error) {
	// eval expression
	return evalStringExpression(s, env.parametersAndVariables()), nil
}
