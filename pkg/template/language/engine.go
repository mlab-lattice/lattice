package language

import (
	"encoding/json"
	"fmt"
	"github.com/mlab-lattice/system/pkg/util/git"
	"strings"
)

const OPERATOR_PREFIX = "$"

// FileResolver interface
type FileResolver interface {
	FileContents(fileName string) ([]byte, error)
}

// Git FileResolver implementation
type GitResolverWrapper struct {
	gitURI             string
	GitResolverContext *git.Context
	GitResolver        *git.Resolver
}

func (gitWrapper GitResolverWrapper) FileContents(fileName string) ([]byte, error) {
	return gitWrapper.GitResolver.FileContents(gitWrapper.GitResolverContext, fileName)
}

// Environment. Template Rendering Environment
type Environment struct {
	currentTemplate *Template
}

// Template Object. Template Rendering Artifact
type Template struct {
	Parent       *Template
	RawMap       map[string]interface{}
	EvaluatedMap map[string]interface{}
	engine       *TemplateEngine
	env          *Environment
}

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
			"$include": &IncludeEvaluator{},
		},
	}

	return engine
}

// ParseTemplate
func (engine *TemplateEngine) ParseTemplate(templateFile string, env *Environment) (*Template, error) {
	rawMap, err := engine.readMapFromFile(env, templateFile)

	if err != nil {
		return nil, err
	}

	var template = &Template{
		RawMap: rawMap,
	}

	if env.currentTemplate != nil {
		template.Parent = env.currentTemplate
	}

	env.currentTemplate = template

	evaluatedMap, err := engine.Eval(rawMap, env)

	if err != nil {
		return nil, err
	}

	template.EvaluatedMap = evaluatedMap.(map[string]interface{})

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
			// return
			if evaluator, isOperator := engine.operatorEvaluators[k]; isOperator {

				return evaluator.EvalOperand(env, engine, v)
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
func (engine *TemplateEngine) evalString(s string, env *Environment) (string, error) {

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

///// Operator Evaluators

type OperatorEvaluator interface {
	EvalOperand(env *Environment, engine *TemplateEngine, operand interface{}) (interface{}, error)
}

// IncludeEvaluator. evaluates
type IncludeEvaluator struct {
}

func (evaluator *IncludeEvaluator) EvalOperand(env *Environment, engine *TemplateEngine, operand interface{}) (interface{}, error) {
	fileName := operand.(string)
	template, err := engine.ParseTemplate(fileName, env)

	if err != nil {
		return nil, err
	} else {
		return template.EvaluatedMap, nil
	}
}
