package language

import (
	"fmt"
)

// OperatorEvaluator

type OperatorEvaluator interface {
	EvalOperand(env *Environment, engine *TemplateEngine, operand interface{}) (interface{}, error)
}

// Used to indicate if the result of the Evaluator is a NOOP
type NOOP int

const NOOP_VAL NOOP = 0

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

	var includeVars map[string]interface{}
	if parameters, hasParams := includeObject["$parameters"]; hasParams {
		var err error
		includeVars, err = evaluator.evaluateParameters(env, engine, parameters.(map[string]interface{}))
		if err != nil {
			return nil, err
		}
	}

	url := includeObject["url"].(string)

	// push the variables into the stack

	currentFrame, err := env.stack.Peek()

	if err != nil {
		return nil, err
	}

	env.stack.Push(&environmentStackFrame{
		variables:    includeVars,
		fileResolver: currentFrame.fileResolver,
	})
	template, err := engine.doParseTemplate(env, url)
	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}

	// Pop
	env.stack.Pop()

	return template.Value, nil

}

func (evaluator *IncludeEvaluator) evaluateParameters(env *Environment, engine *TemplateEngine, parameters map[string]interface{}) (map[string]interface{}, error) {

	variables := make(map[string]interface{})
	for name, rawVal := range parameters {
		paramVal, err := engine.Eval(rawVal, env)
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

func (evaluator *VariablesEvaluator) EvalOperand(env *Environment, engine *TemplateEngine, operand interface{}) (interface{}, error) {
	variablesMap := operand.(map[string]interface{})

	// get current stack frame
	currentFrame, err := env.stack.Peek()

	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}
	for name, rawVal := range variablesMap {
		val, err := engine.Eval(rawVal, env)
		if err != nil {
			return nil, err
		}
		currentFrame.variables[name] = val
	}

	// NOOP
	return NOOP_VAL, nil

}

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
