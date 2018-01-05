package language

import (
	"fmt"
)

// OperatorEvaluator operator evaluators used by the engine to evaluate special operators

type OperatorEvaluator interface {
	eval(value interface{}, env *environment) (interface{}, error)
}

// Used to indicate if the result of the Evaluator is a NOOP
type Void int

const void Void = 0

// IncludeEvaluator. evaluates $include
type IncludeEvaluator struct {
}

func (evaluator *IncludeEvaluator) eval(value interface{}, env *environment) (interface{}, error) {

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

	var includeParameters map[string]interface{}
	if parameters, hasParams := includeObject["parameters"]; hasParams {
		var err error
		includeParameters, err = evaluator.evaluateParameters(parameters.(map[string]interface{}), env)
		if err != nil {
			return nil, err
		}
	}

	url := includeObject["url"].(string)

	return env.engine.includeUrl(url, includeParameters, env)
}

// evaluateParameters evaluates parameters to passed for the $include
func (evaluator *IncludeEvaluator) evaluateParameters(parameters map[string]interface{}, env *environment) (map[string]interface{}, error) {

	variables := make(map[string]interface{})
	for name, rawVal := range parameters {
		paramVal, err := env.engine.eval(rawVal, env)
		if err != nil {
			return nil, err
		}
		variables[name] = paramVal
	}

	return variables, nil
}

// VariablesEvaluator. evaluates $variables
type VariablesEvaluator struct {
}

// eval
func (evaluator *VariablesEvaluator) eval(value interface{}, env *environment) (interface{}, error) {
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
		val, err := env.engine.eval(rawVal, env)
		if err != nil {
			return nil, err
		}
		currentFrame.parameters[name] = val
	}

	// void
	return void, nil

}

// ParametersEvaluator. evaluates $parameters
type ParametersEvaluator struct {
}

// eval
func (evaluator *ParametersEvaluator) eval(value interface{}, env *environment) (interface{}, error) {
	paramMap := value.(map[string]interface{})

	for name, paramDef := range paramMap {
		err := evaluator.processParam(name, paramDef.(map[string]interface{}), env)
		if err != nil {
			return nil, err
		}
	}

	// void
	return void, nil

}

// processParam process/validate parameters
func (evaluator *ParametersEvaluator) processParam(name string, paramDef map[string]interface{}, env *environment) error {
	// get current stack frame
	currentFrame, err := env.stack.Peek()

	if err != nil {
		return err
	}
	// validate required
	if isRequired, requiredIsSet := paramDef["required"]; requiredIsSet && isRequired.(bool) {
		if _, paramIsSet := currentFrame.parameters[name]; !paramIsSet {
			return fmt.Errorf("parameter %s is required", name)
		}
	}

	// default param as needed
	if defaultValue, hasDefault := paramDef["required"]; hasDefault {
		if _, paramIsSet := currentFrame.parameters[name]; !paramIsSet {
			currentFrame.parameters[name] = defaultValue
		}
	}

	return nil

}
