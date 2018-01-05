package language

import (
	"encoding/json"
	"fmt"
	"github.com/mlab-lattice/system/pkg/util/git"
	"path"
	"strings"
)

const operatorPrefix = "$"

/**

Design:

Main interface: EvalFromURL(url, parameters, git.Options)

* EvalFromURL(url, parameters, options)
   - create env
   - includeUrl(url, parameters, env)

* includeUrl(url, env)
   - parse url
   - determine repository/file-path
   - readFileBytes
   - unmarshal
   - push stack frame
   - eval
   - pop

* $include
   - eval parameters
   - includeUrl(url, newParameters, env)

*/

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
func (engine *TemplateEngine) EvalFromURL(url string, parameters map[string]interface{}, gitOptions *git.Options) (interface{}, error) {
	env := newEnvironment(engine, gitOptions)

	return engine.includeUrl(url, parameters, env)
}

// includeUrl
func (engine *TemplateEngine) includeUrl(url string, parameters map[string]interface{}, env *environment) (interface{}, error) {

	// parse url
	urlInfo, err := parseTemplateUrl(url, env)

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

	// Push a frame to the stack
	env.stack.Push(&environmentStackFrame{
		parameters:     parameters,
		fileRepository: fileRepository,
		filePath:       urlInfo.filePath,
	})

	rawValue, err := engine.readFileBytes(filePath, env)

	if err != nil {
		return nil, err
	}

	val, err := engine.eval(rawValue, env)

	if err != nil {
		return nil, err
	}

	// Pop
	env.stack.Pop()

	return val, nil
}

// eval evaluates the zsfgh
func (engine *TemplateEngine) eval(v interface{}, env *environment) (interface{}, error) {

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
	// get current frame
	currentFrame, err := env.stack.Peek()
	if err != nil {
		return nil, err
	}

	// eval expression
	return evalStringExpression(s, currentFrame.parameters)
}

// includeFile includes the file and returns a map of values
func (engine *TemplateEngine) readFileBytes(filePath string, env *environment) (map[string]interface{}, error) {

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

	val, err := engine.eval(rawMap, env)

	if err != nil {
		return nil, err

	}

	return val.(map[string]interface{}), nil
}

// unmarshalBytes unmarshal the bytes specified based on the the file name
func (engine *TemplateEngine) unmarshalBytes(bytes []byte, filePath string, env *environment) (map[string]interface{}, error) {

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
