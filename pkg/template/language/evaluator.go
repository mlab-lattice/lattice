package language

import (
	"fmt"
)

// OperatorEvaluator

type OperatorEvaluator interface {
	eval(value interface{}, env *Environment) (interface{}, error)
}

// Used to indicate if the result of the Evaluator is a NOOP
type NOOP int

const NOOP_VAL NOOP = 0

// IncludeEvaluator. evaluates $include
type IncludeEvaluator struct {
}

func (evaluator *IncludeEvaluator) eval(value interface{}, env *Environment) (interface{}, error) {

	// construct the include object. We allow the include to be an object or a string.
	// string will be converted to to {url: val}
	var includeObject map[string]interface{}

	if _, isMap := value.(map[string]interface{}); isMap {
		includeObject = value.(map[string]interface{})
	} else if _, isString := value.(string); isString {
		includeObject = map[string]interface{}{
			"url": value,
		}
	} else {
		return nil, fmt.Errorf("Invalid $include %s", includeObject)
	}

	// validate include object
	if _, hasUrl := includeObject["url"]; !hasUrl {
		return nil, fmt.Errorf("$include has no url %s", includeObject)
	}

	//evaluate parameters if present

	var includeVars map[string]interface{}
	if parameters, hasParams := includeObject["parameters"]; hasParams {
		var err error
		includeVars, err = evaluator.evaluateParameters(parameters.(map[string]interface{}), env)
		if err != nil {
			return nil, err
		}
	}

	url := includeObject["url"].(string)

	template, err := env.engine.doParseTemplate(url, includeVars, env)
	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}

	return template.Value, nil

}

func (evaluator *IncludeEvaluator) evaluateParameters(parameters map[string]interface{}, env *Environment) (map[string]interface{}, error) {

	variables := make(map[string]interface{})
	for name, rawVal := range parameters {
		paramVal, err := env.engine.Eval(rawVal, env)
		if err != nil {
			return nil, err
		}
		variables[name] = paramVal
	}

	return variables, nil
}

/**********************************************************************************************************************/
// VariablesEvaluator. evaluates
type VariablesEvaluator struct {
}

func (evaluator *VariablesEvaluator) eval(value interface{}, env *Environment) (interface{}, error) {
	variablesMap := value.(map[string]interface{})

	// get current stack frame
	currentFrame, err := env.stack.Peek()

	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}
	for name, rawVal := range variablesMap {
		val, err := env.engine.Eval(rawVal, env)
		if err != nil {
			return nil, err
		}
		currentFrame.variables[name] = val
	}

	// NOOP
	return NOOP_VAL, nil

}

// ParametersEvaluator. evaluates $parameters
type ParametersEvaluator struct {
}

func (evaluator *ParametersEvaluator) eval(value interface{}, env *Environment) (interface{}, error) {
	paramMap := value.(map[string]interface{})

	for name, paramDef := range paramMap {
		err := evaluator.processParam(name, paramDef.(map[string]interface{}), env)
		if err != nil {
			return nil, err
		}
	}

	// NOOP
	return NOOP_VAL, nil

}

func (evaluator *ParametersEvaluator) processParam(name string, paramDef map[string]interface{}, env *Environment) error {
	// get current stack frame
	currentFrame, err := env.stack.Peek()

	if err != nil {
		return err
	}
	// validate required
	if isRequired, requiredIsSet := paramDef["required"]; requiredIsSet && isRequired.(bool) {
		if _, paramIsSet := currentFrame.variables[name]; !paramIsSet {
			return fmt.Errorf("parameter %s is required", name)
		}
	}

	// default param as needed
	if defaultValue, hasDefault := paramDef["required"]; hasDefault {
		if _, paramIsSet := currentFrame.variables[name]; !paramIsSet {
			currentFrame.variables[name] = defaultValue
		}
	}

	return nil

}
