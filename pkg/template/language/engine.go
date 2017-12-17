package language

import (
	"encoding/json"
	"fmt"
	"strings"
)

const OPERATOR_PREFIX = "$"


// Template Object. Template Rendering Artifact
type Template struct {
	Value map[string]interface{}
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

// ParseTemplate
func (engine *TemplateEngine) ParseTemplate(url string, variables map[string]interface{}, fileResolver FileResolver) (*Template, error) {
	env := newEnvironment()
	// Push Variables to stack
	env.stack.Push(&environmentStackFrame{
		variables:    variables,
		fileResolver: fileResolver,
	})

	return engine.doParseTemplate(env, url)
}

func (engine *TemplateEngine) doParseTemplate(env *Environment, url string) (*Template, error) {

	rawMap, err := engine.readMapFromFile(env, url)

	if err != nil {
		return nil, err
	}

	val, err := engine.Eval(rawMap, env)

	if err != nil {
		return nil, err
	}

	var template = &Template{Value: val.(map[string]interface{})}

	if err != nil {
		return nil, err
	}

	return template, nil
}

// eval resolves a single json value. i.e. deals with special values such as $include
func (engine *TemplateEngine) Eval(v interface{}, env *Environment) (interface{}, error) {

	// check value type and use proper eval method
	if valMap, isMap := v.(map[string]interface{}); isMap { // Maps
		return engine.evalMap(valMap, env)

	} else if valArr, isArray := v.([]interface{}); isArray { // Arrays
		return engine.evalArray(valArr, env)

	} else if stringVal, isString := v.(string); isString { // Strings
		return engine.evalString(stringVal, env)

	} else { // Default, just return the value as is
		return v, nil
	}

}

// evaluates a map of objects
func (engine *TemplateEngine) evalMap(mapVal map[string]interface{}, env *Environment) (interface{}, error) {
	result := make(map[string]interface{})
	for k, v := range mapVal {
		var err error
		// check if the key is an operator
		if strings.HasPrefix(k, OPERATOR_PREFIX) {

			if evaluator, isOperator := engine.operatorEvaluators[k]; isOperator {

				evalResult, err := evaluator.EvalOperand(env, engine, v)
				if err != nil {
					return nil, err
				} else if evalResult == NOOP_VAL { // NOOP case, just skip
					continue
				} else {
					return evalResult, nil
				}

			}
		}

		result[k], err = engine.Eval(v, env)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

// evaluates an array of objects
func (engine *TemplateEngine) evalArray(arrayVal []interface{}, env *Environment) ([]interface{}, error) {
	result := make([]interface{}, len(arrayVal))
	for i, v := range arrayVal {
		var err error
		result[i], err = engine.Eval(v, env)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

// evalutes a string val
func (engine *TemplateEngine) evalString(s string, env *Environment) (interface{}, error) {
	// get current frame
	currentFrame, err := env.stack.Peek()
	if err != nil {
		return nil, err
	}

	// quick hack to evaluate variable references
	if strings.HasPrefix(s, "${") {
		varName := strings.TrimSuffix(strings.TrimPrefix(s, "${"), "}")
		return currentFrame.variables[varName], nil
	}

	return s, nil
}

func (engine *TemplateEngine) readMapFromFile(env *Environment, fileName string) (map[string]interface{}, error) {

	// get current frame
	currentFrame, err := env.stack.Peek()

	if err != nil {
		return nil, err
	}

	jsonBytes, err := currentFrame.fileResolver.FileContents(fileName)
	if err != nil {
		return nil, err
	}

	result := make(map[string]interface{})

	// unmarshal file contents based on file type. Only .json is supported atm

	if strings.HasSuffix(fileName, ".json") {
		err = json.Unmarshal(jsonBytes, &result)

		if err != nil {
			return nil, err
		} else {
			return result, nil
		}
	} else {
		return nil, error(fmt.Errorf("Unsupported file %s", fileName))
	}

}
