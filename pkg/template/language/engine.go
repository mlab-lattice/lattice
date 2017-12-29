package language

import (
	"encoding/json"
	"fmt"
	"path"
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
func (engine *TemplateEngine) ParseTemplate(url string, variables map[string]interface{}) (*Template, error) {
	env := newEnvironment(engine)

	return engine.doParseTemplate(url, variables, env)
}

func (engine *TemplateEngine) doParseTemplate(url string, variables map[string]interface{}, env *Environment) (*Template, error) {

	// parse url
	urlInfo, err := parseTemplateUrl(url)

	if err != nil {
		return nil, err
	}

	fileRepository := urlInfo.fileRepository

	// if its not a new file repository then use the current one
	if fileRepository == nil {
		currentFrame, err := env.stack.Peek()
		if err != nil {
			return nil, err
		}
		fileRepository = currentFrame.fileRepository
	}

	// construct file path relative to the parent file if specified

	filePath := path.Join(env.currentDir(), urlInfo.filePath)

	// Push Variables to stack
	env.stack.Push(&environmentStackFrame{
		variables:      variables,
		fileRepository: fileRepository,
		filePath:       urlInfo.filePath,
	})

	rawMap, err := engine.includeFile(filePath, env)

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

	// Pop
	env.stack.Pop()

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

				evalResult, err := evaluator.eval(v, env)
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

// includeFile includes the file and returns a map of values
func (engine *TemplateEngine) includeFile(filePath string, env *Environment) (map[string]interface{}, error) {

	fmt.Printf("Including file %s\n", filePath)
	currentFrame, err := env.stack.Peek()
	if err != nil {
		return nil, err
	}

	bytes, err := currentFrame.fileRepository.getFileContents(filePath)
	if err != nil {
		return nil, err
	}

	rawMap, err := engine.unmarshalBytes(bytes, filePath, env)

	if err != nil {
		return nil, err
	}

	val, err := engine.Eval(rawMap, env)

	if err != nil {
		return nil, err

	}

	return val.(map[string]interface{}), nil
}

// unmarshalBytes unmarshal the bytes specified based on the the file name
func (engine *TemplateEngine) unmarshalBytes(bytes []byte, filePath string, env *Environment) (map[string]interface{}, error) {

	// unmarshal file contents based on file type. Only .json is supported atm

	result := make(map[string]interface{})

	if strings.HasSuffix(filePath, ".json") {
		err := json.Unmarshal(bytes, &result)

		if err != nil {
			return nil, err
		} else {
			return result, nil
		}
	} else {
		return nil, error(fmt.Errorf("Unsupported file %s", filePath))
	}

}
