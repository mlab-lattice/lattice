// Package language provides json template language.
package language

/*
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

import (
	"fmt"
	"github.com/mlab-lattice/lattice/pkg/util/git"
)

// Options evaluation options
type Options struct {
	WorkDirectory string       // work directory passed for git resolver
	GitOptions    *git.Options // git options to be passed for git resolver
}

func CreateOptions(workDirectory string, gitOptions *git.Options) (*Options, error) {

	if workDirectory == "" {
		return nil, fmt.Errorf("must supply workDirectory")
	}

	options := &Options{
		WorkDirectory: workDirectory,
		GitOptions:    gitOptions,
	}

	return options, nil
}

// TemplateEngine the main class to be used for parsing/evaluating templates
type TemplateEngine struct {
	operatorConfigs []*operatorConfig // array of operator configs. Order of elements here is used as the
	// evaluation order of operators within a map
	operatorMap map[string]*operatorConfig // internal operator evaluator registry.

}

// NewEngine constructs new engine object
func NewEngine() *TemplateEngine {

	operatorConfigs := []*operatorConfig{
		{
			key:                   "$parameters",
			evaluator:             &ParametersEvaluator{},
			appendToPropertyStack: true,
		},
		{
			key:                   "$variables",
			evaluator:             &VariablesEvaluator{},
			appendToPropertyStack: true,
		},
		{
			key:                   "$include",
			evaluator:             &IncludeEvaluator{},
			appendToPropertyStack: false,
		},
	}

	// construct the operator map based on configs
	operatorMap := make(map[string]*operatorConfig)
	for _, opConf := range operatorConfigs {
		operatorMap[opConf.key] = opConf
	}

	// create the engine
	engine := &TemplateEngine{
		operatorConfigs: operatorConfigs,
		operatorMap:     operatorMap,
	}

	return engine
}

// EvalFromURL evaluates the template from the specified url with the specified parameters and options
func (engine *TemplateEngine) EvalFromURL(url string, parameters map[string]interface{}, options *Options) (*Result, error) {

	// make parameters if not set
	if parameters == nil {
		parameters = make(map[string]interface{})
	}

	env := newEnvironment(engine, options)
	val, err := engine.include(url, parameters, env)
	if err != nil {
		return nil, err
	}

	return newResult(val, env), nil
}

// Eval evaluates a single object
func (engine *TemplateEngine) Eval(o interface{}, parameters map[string]interface{},
	options *Options) (*Result, error) {

	// make parameters if not set
	if parameters == nil {
		parameters = make(map[string]interface{})
	}

	// create env and push parameters to the stack
	env := newEnvironment(engine, options)
	env.push(nil, parameters, make(map[string]interface{}))
	// defer pop
	defer env.pop()

	// call eval with env
	val, err := engine.eval(o, env)

	if err != nil {
		return nil, err
	}

	return newResult(val, env), err
}

// eval evaluates the specified object
func (engine *TemplateEngine) eval(o interface{}, env *environment) (interface{}, error) {

	if valMap, isMap := o.(map[string]interface{}); isMap { // Maps
		return engine.evalMap(valMap, env)

	} else if valArr, isArray := o.([]interface{}); isArray { // Arrays
		return engine.evalArray(valArr, env)

	} else if stringVal, isString := o.(string); isString { // Strings
		return engine.evalString(stringVal, env)

	}
	// Default, just return the value as is
	return o, nil
}

// include includes and evaluates the template file specified in the url
func (engine *TemplateEngine) include(url string, parameters map[string]interface{}, env *environment) (interface{}, error) {
	// resolve url
	resource, err := resolveURL(url, env)

	if err != nil {
		return nil, err
	}

	// init variables
	variables := make(map[string]interface{})

	// push !
	env.push(resource, parameters, variables)

	// defer a pop to ensure that the stack is popped  before
	defer env.pop()

	// evaluate data of the template
	return engine.eval(resource.data, env)
}

// evalMap evaluates a map of objects
func (engine *TemplateEngine) evalMap(m map[string]interface{}, env *environment) (interface{}, error) {
	// init result
	result := make(map[string]interface{})

	// eval operators in map
	err := engine.evalOperatorsInMap(result, m, env)

	if err != nil {
		return nil, err
	}

	// eval properties
	err = engine.evalPropertiesInMap(result, m, env)

	if err != nil {
		return nil, err
	}

	return result, nil
}

// evalOperatorsInMap helper method for evalMap
func (engine *TemplateEngine) evalOperatorsInMap(result map[string]interface{}, m map[string]interface{},
	env *environment) error {

	// first, evaluate operators based on their priorities

	for _, operator := range engine.operatorConfigs {
		if operand, operatorExists := m[operator.key]; operatorExists {

			// push the the current operator to the property stack
			currentPropertyPath := env.getCurrentPropertyPath()
			if operator.appendToPropertyStack {
				currentPropertyPath = env.pushProperty(operator.key)
			}

			evalResult, err := operator.evaluator.eval(operand, env)

			// pop property if we pushed it
			if operator.appendToPropertyStack {
				env.popProperty()
			}

			if err != nil { // return error
				return wrapWithPropertyEvalError(err, currentPropertyPath, env)
			}
			// if the result is nil, this indicates that the evaluator has processed the operator and that the engine
			// should continue
			if evalResult == nil {
				continue
			}

			// evaluator has return a value, we expect this value to be a map and we just merge it with the existing
			// result
			resultMap, isMap := evalResult.(map[string]interface{})

			if !isMap {
				badOperatorResultError := fmt.Errorf("Bad return value for evaluator %v. "+
					"Result is not a map", operator)
				return wrapWithPropertyEvalError(badOperatorResultError, currentPropertyPath, env)
			}
			// stuff map with the val result
			for k, v := range resultMap {
				result[k] = v
			}
		}
	}

	return nil
}

// evalPropertiesInMap helper method for evalMap
func (engine *TemplateEngine) evalPropertiesInMap(result map[string]interface{}, m map[string]interface{},
	env *environment) error {
	// eval the rest of the map
	for k, v := range m {

		// skip operators since we have evaluated them already
		if _, isOperator := engine.operatorMap[k]; isOperator {
			continue
		}

		// regular property. Push to the property stack
		currentPropertyPath := env.pushProperty(k)

		var err error

		// eval
		result[k], err = engine.eval(v, env)

		// pop property
		env.popProperty()
		if err != nil {
			return wrapWithPropertyEvalError(err, currentPropertyPath, env)
		}
	}

	return nil
}

// evalArray evaluates an array of objects
func (engine *TemplateEngine) evalArray(arr []interface{}, env *environment) ([]interface{}, error) {
	result := make([]interface{}, len(arr))
	for i, v := range arr {
		// construct a property for the array element as array.index e.g. "items.0"
		arrayElementProperty := fmt.Sprintf("%v", i)
		currentPropertyPath := env.pushProperty(arrayElementProperty)
		var err error
		result[i], err = engine.eval(v, env)
		env.popProperty()
		if err != nil {
			return nil, wrapWithPropertyEvalError(err, currentPropertyPath, env)
		}
	}

	return result, nil
}

// evaluates a string
func (engine *TemplateEngine) evalString(s string, env *environment) (interface{}, error) {
	// eval expression
	return evalStringExpression(s, env.parametersAndVariables()), nil
}
