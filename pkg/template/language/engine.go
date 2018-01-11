package language

import (
	"github.com/mlab-lattice/system/pkg/util/git"
)

/**

Design:


* Eval(o, params, options)
   - create env
   - eval(o, params, env)

* eval(o, params, env)
 - string: evalString(o, params, env)
 - map: evalMap(o, params, env)
 - template: evalTemplate(o, params, env)
 - default: return o


* env:
  - stack<stackFrame>
  - stackFrame:
     - parameters
     - variables
     - baseUrl

* "$include" : {
    "url": "URL",
    "parameters": {
       <name: val>,
    }
  }

   - validate parameters/apply default values
   - engine.include(url, newParameters, env)

* "$variables": {
     <name: val>,
   }

  - env.variables[name] = engine.eval(variable)

* "$parameters": {
     "<name>": {
       "required": <bool>,
       "default": <object>
     }
   }


TODO:
 * Ordering in map evaluation.
    - Order
    - maybe: $parameters, $variables, remaining
    - Or order by line numbers in template itself if possible

 * Ordering in $variable evaluation: variable def using a value of a another variable
    - order of line numbers
    - OR maybe make variable def a list instead of a map

 * what if a variable and a parameter has the same name? who wins?
 * Keep line numbers, maybe build a syntax tree where we have the preserve the property path and line number

 * Git urls: Which form do we want to allow?
  - If we allow ssh:// , what does that imply?
  - I am thinking allow file:// and git://
  - Whatever format we use, it should be allowed by the git/go-git as is

* Allow escaping in string interpolations

*/

type Options struct {
	GitOptions *git.Options
}

// TemplateEngine
type TemplateEngine struct {
	operatorEvaluators map[string]OperatorEvaluator
	operatorEvalOrder  []string
}

// NewEngine
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

// EvalFromURL
func (engine *TemplateEngine) EvalFromURL(url string, parameters map[string]interface{}, options *Options) (map[string]interface{}, error) {
	env := newEnvironment(engine, options)
	result, err := engine.include(url, parameters, env)
	if err != nil {
		return nil, err
	}

	return result.(map[string]interface{}), nil
}

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

	// init variables
	variables := make(map[string]interface{})

	// push !
	env.push(resource.baseUrl, parameters, variables)

	if err != nil {
		return nil, err
	}

	// evaluate data of the template
	val, err := engine.eval(resource.data, env)

	if err != nil {
		return nil, err
	}

	// pop
	env.pop()

	return val, nil
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
			} else if evalResult == void { // NOOP case, just skip
				continue
			} else {
				return evalResult, nil
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
	return evalStringExpression(s, env.parametersAndVariables())
}
