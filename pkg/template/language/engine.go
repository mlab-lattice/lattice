package language

import (
	"strings"

	"github.com/mlab-lattice/system/pkg/util/git"
)

const operatorPrefix = "$"

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
     <name: val>,
   }




TODO:
 - allow escaping in string interpolations
 - Ordering in map evaluation.
    - maybe: $parameters, $variables, remaining
    - Or order by line numbers in template itself if possible

 - Ordering in $variable evaluation: variable def using a value of a another variable
    - order of line numbers
    - OR maybe make variable def a list instead of a map

 - what if a variable and a parameter has the same name? who wins?
 - Keep line numbers

*/

type Options struct {
	gitOptions *git.Options
}

// TemplateEngine
type TemplateEngine struct {
	operatorEvaluators map[string]OperatorEvaluator
}

// NewEngine
func NewEngine() *TemplateEngine {

	engine := &TemplateEngine{
		operatorEvaluators: map[string]OperatorEvaluator{
			"$include":    &IncludeEvaluator{},
			"$variables":  &VariablesEvaluator{},
			"$parameters": &ParametersEvaluator{},
		},
	}

	return engine
}

// EvalFromURL
func (engine *TemplateEngine) EvalFromURL(url string, parameters map[string]interface{}, options *Options) (interface{}, error) {
	env := newEnvironment(engine, options)
	return engine.include(url, parameters, env)
}

// eval evaluates the
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

func (engine *TemplateEngine) include(url string, parameters map[string]interface{}, env *environment) (interface{}, error) {
	// resolve url
	resource, err := resolveUrl(url, env)

	if err != nil {
		return nil, err
	}

	variables := make(map[string]interface{})

	// push !
	env.push(resource.baseUrl, parameters, variables)

	if err != nil {
		return nil, err
	}

	val, err := engine.eval(resource.data, env)

	if err != nil {
		return nil, err
	}

	// pop
	env.pop()

	return val, nil
}

// evaluates a map of objects
func (engine *TemplateEngine) evalMap(mapVal map[string]interface{}, env *environment) (interface{}, error) {
	result := make(map[string]interface{})
	for k, v := range mapVal {
		var err error
		// check if the key is an operator
		if strings.HasPrefix(k, operatorPrefix) {

			if evaluator, isOperator := engine.operatorEvaluators[k]; isOperator {

				evalResult, err := evaluator.eval(v, env)
				if err != nil {
					return nil, err
				} else if evalResult == void { // NOOP case, just skip
					continue
				} else {
					return evalResult, nil
				}

			}
		}

		result[k], err = engine.eval(v, env)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

// evaluates an array of objects
func (engine *TemplateEngine) evalArray(arrayVal []interface{}, env *environment) ([]interface{}, error) {
	result := make([]interface{}, len(arrayVal))
	for i, v := range arrayVal {
		var err error
		result[i], err = engine.eval(v, env)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

// evalutes a string val
func (engine *TemplateEngine) evalString(s string, env *environment) (interface{}, error) {
	// eval expression
	return evalStringExpression(s, env.parametersAndVariables())
}
