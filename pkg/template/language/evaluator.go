package language

import (
	"fmt"
)

// OperatorEvaluator operator evaluators used by the engine to evaluate special operators

type OperatorEvaluator interface {
	eval(o interface{}, env *environment) (interface{}, error)
}

// Used to indicate if the result of the Evaluator is a Void
type Void int

const void Void = 0

// IncludeEvaluator. evaluates $include
type IncludeEvaluator struct {
}

func (evaluator *IncludeEvaluator) eval(o interface{}, env *environment) (interface{}, error) {

	// construct the include object. We allow the include to be an object or a string.
	// string will be converted to to {url: val}
	var includeObject map[string]interface{}
	if _, isMap := o.(map[string]interface{}); isMap {
		includeObject = o.(map[string]interface{})
	} else if _, isString := o.(string); isString {
		includeObject = map[string]interface{}{
			"url": o,
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
	if includeParamsVal, hasParams := includeObject["parameters"]; hasParams {
		var err error
		params, err := env.engine.eval(includeParamsVal, env)
		if err != nil {
			return nil, err
		}

		includeParameters = params.(map[string]interface{})
	}

	url := includeObject["url"].(string)

	return env.engine.include(url, includeParameters, env)
}

// VariablesEvaluator. evaluates $variables
type VariablesEvaluator struct {
}

// eval
func (evaluator *VariablesEvaluator) eval(o interface{}, env *environment) (interface{}, error) {
	variables, err := env.engine.eval(o, env)
	if err != nil {
		return nil, err
	}

	env.currentFrame().variables = variables.(map[string]interface{})

	// void
	return void, nil

}

// ParametersEvaluator. evaluates $parameters
type ParametersEvaluator struct {
}

// eval
func (evaluator *ParametersEvaluator) eval(o interface{}, env *environment) (interface{}, error) {
	paramMap := o.(map[string]interface{})

	for name, paramDef := range paramMap {
		err := evaluator.processInputParameter(name, paramDef.(map[string]interface{}), env)
		if err != nil {
			return nil, err
		}
	}

	// void
	return void, nil

}

// processInputParameter process/validate template parameters
func (evaluator *ParametersEvaluator) processInputParameter(name string, paramDef map[string]interface{}, env *environment) error {
	parameters := env.currentFrame().parameters
	// validate required
	if isRequired, requiredIsSet := paramDef["required"]; requiredIsSet && isRequired.(bool) {
		if _, paramIsSet := parameters[name]; !paramIsSet {
			return fmt.Errorf("parameter %s is required", name)
		}
	}
	// default param as needed
	if defaultValue, hasDefault := paramDef["required"]; hasDefault {
		if _, paramIsSet := parameters[name]; !paramIsSet {
			parameters[name] = defaultValue
		}
	}

	return nil

}
