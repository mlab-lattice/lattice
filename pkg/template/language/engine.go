package language

import (
	"encoding/json"
	"fmt"
	"github.com/mlab-lattice/system/pkg/util/git"
	"strings"
)

const OPERATOR_PREFIX = "$"

/**********************************************************************************************************************/
// FileResolver interface
type FileResolver interface {
	FileContents(fileName string) ([]byte, error)
}

/**********************************************************************************************************************/
// Git FileResolver implementation
type GitResolverWrapper struct {
	gitURI             string
	GitResolverContext *git.Context
	GitResolver        *git.Resolver
}

func (gitWrapper GitResolverWrapper) FileContents(fileName string) ([]byte, error) {
	return gitWrapper.GitResolver.FileContents(gitWrapper.GitResolverContext, fileName)
}

/**********************************************************************************************************************/
// Environment. Template Rendering Environment
type Environment struct {
	currentTemplate *Template
	Variables       map[string]interface{}
}

// NewEnvironment
func NewEnvironment() *Environment {

	env := &Environment{
		Variables: make(map[string]interface{}),
	}

	return env
}

/**********************************************************************************************************************/
// Template Object. Template Rendering Artifact
type Template struct {
	Parent       *Template
	RawMap       map[string]interface{}
	EvaluatedMap map[string]interface{}
	engine       *TemplateEngine
	env          *Environment
}

/**********************************************************************************************************************/
// TemplateEngine
type TemplateEngine struct {
	FileResolver       FileResolver
	operatorEvaluators map[string]OperatorEvaluator
}

// NewEngine
func NewEngine(fileResolver FileResolver) *TemplateEngine {

	engine := &TemplateEngine{
		FileResolver: fileResolver,
		operatorEvaluators: map[string]OperatorEvaluator{
			"$include":    &IncludeEvaluator{},
			"$variables":  &VariablesEvaluator{},
			"$parameters": &ParametersEvaluator{},
		},
	}

	return engine
}

// ParseTemplate
func (engine *TemplateEngine) ParseTemplate(url string, env *Environment) (*Template, error) {
	rawMap, err := engine.readMapFromFile(env, url)

	if err != nil {
		return nil, err
	}

	var template = &Template{
		RawMap: rawMap,
	}

	env.currentTemplate = template

	evaluatedMap, err := engine.Eval(rawMap, env)

	if err != nil {
		return nil, err
	}

	template.EvaluatedMap = evaluatedMap.(map[string]interface{})

	return template, nil
}

func (engine *TemplateEngine) parseChildTemplate(url string, parentTemplate *Template, env *Environment) (*Template, error) {
	childTemplate, err := engine.ParseTemplate(url, env)

	if err != nil {
		return nil, err
	}

	childTemplate.Parent = parentTemplate

	return childTemplate, nil
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

	// quick hack to evaluate variable references
	if strings.HasPrefix(s, "${") {
		varName := strings.TrimSuffix(strings.TrimPrefix(s, "${"), "}")
		return env.Variables[varName], nil
	}

	return s, nil
}

// reads/consolidates json map form a file.
func (engine *TemplateEngine) readConsolidatedJsonMapFromFile(env *Environment, fileName string) (map[string]interface{}, error) {

	jsonBytes, err := engine.FileResolver.FileContents(fileName)
	if err != nil {
		return nil, err
	}
	result := make(map[string]interface{})
	err = json.Unmarshal(jsonBytes, &result)

	if err != nil {
		return nil, err
	}

	// resolve json and bytes
	engine.Eval(result, env)

	return result, nil
}

func (engine *TemplateEngine) readMapFromFile(env *Environment, fileName string) (map[string]interface{}, error) {

	jsonBytes, err := engine.FileResolver.FileContents(fileName)
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

/**********************************************************************************************************************/
// OperatorEvaluator

type OperatorEvaluator interface {
	EvalOperand(env *Environment, engine *TemplateEngine, operand interface{}) (interface{}, error)
}

/**********************************************************************************************************************/
// Used to indicate if the result of the Evaluator is a NOOP
type NOOP int

const NOOP_VAL NOOP = 0

/**********************************************************************************************************************/
// IncludeEvaluator. evaluates
type IncludeEvaluator struct {
}

func (evaluator *IncludeEvaluator) EvalOperand(env *Environment, engine *TemplateEngine, operand interface{}) (interface{}, error) {

	// construct the include object. We allow the include to be an object or a string.
	// string will be converted to to {url: val}
	var includeObject map[string]interface{}

	if _, isMap := operand.(map[string]interface{}); isMap {
		includeObject = operand.(map[string]interface{})
	} else if _, isString := operand.(string); isString {
		includeObject = map[string]interface{}{
			"url": operand,
		}
	} else {
		return nil, fmt.Errorf("Invalid $include %s", includeObject)
	}

	// validate include object
	if _, hasUrl := includeObject["url"]; !hasUrl {
		return nil, fmt.Errorf("$include has no url %s", includeObject)
	}

	//evaluate parameters if present
	childEnv := NewEnvironment()

	if parameters, hasParams := includeObject["$parameters"]; hasParams {
		err := evaluator.evaluateParameters(childEnv, engine, parameters.(map[string]interface{}))
		if err != nil {
			return nil, err
		}
	}

	url := includeObject["url"].(string)
	template, err := engine.parseChildTemplate(url, env.currentTemplate, childEnv)

	if err != nil {
		return nil, err
	} else {
		return template.EvaluatedMap, nil
	}
}

func (evaluator *IncludeEvaluator) evaluateParameters(env *Environment, engine *TemplateEngine, parameters map[string]interface{}) error {

	for name, rawVal := range parameters {
		paramVal, err := engine.Eval(rawVal, env)
		if err != nil {
			return err
		}
		env.Variables[name] = paramVal
	}

	return nil
}

/**********************************************************************************************************************/
// VariablesEvaluator. evaluates
type VariablesEvaluator struct {
}

func (evaluator *VariablesEvaluator) EvalOperand(env *Environment, engine *TemplateEngine, operand interface{}) (interface{}, error) {
	variablesMap := operand.(map[string]interface{})

	for name, rawVal := range variablesMap {
		val, err := engine.Eval(rawVal, env)
		if err != nil {
			return nil, err
		}
		env.Variables[name] = val
	}

	// NOOP
	return NOOP_VAL, nil

}

/**********************************************************************************************************************/
// ParametersEvaluator. evaluates
type ParametersEvaluator struct {
}

func (evaluator *ParametersEvaluator) EvalOperand(env *Environment, engine *TemplateEngine, operand interface{}) (interface{}, error) {
	paramMap := operand.(map[string]interface{})

	for name, paramDef := range paramMap {
		err := evaluator.processParam(env, name, paramDef.(map[string]interface{}))
		if err != nil {
			return nil, err
		}
	}

	// NOOP
	return NOOP_VAL, nil

}

func (evaluator *ParametersEvaluator) processParam(env *Environment, name string, paramDef map[string]interface{}) error {

	// validate required
	if isRequired, requiredIsSet := paramDef["required"]; requiredIsSet && isRequired.(bool) {
		if _, paramIsSet := env.Variables[name]; !paramIsSet {
			return fmt.Errorf("parameter %s is required", name)
		}
	}

	// default param as needed
	if defaultValue, hasDefault := paramDef["required"]; hasDefault {
		if _, paramIsSet := env.Variables[name]; !paramIsSet {
			env.Variables[name] = defaultValue
		}
	}

	return nil

}
