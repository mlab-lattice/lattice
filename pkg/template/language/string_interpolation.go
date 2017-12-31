package language

import (
	"regexp"
)

// Contains functions needed for evaluating string expressions

// used for matching a single variable reference expression. i.e. ${var}
var singleVarRegex = regexp.MustCompile(`^\$\{([a-zA-Z_$][a-zA-Z_$0-9]*)\}$`)

// evalStringExpression
func evalStringExpression(expression string, variables map[string]interface{}) (interface{}, error) {

	if isSingleVariableExpression(expression) {
		return evalSingleVariableExpression(expression, variables)
	}

	// otherwise just return the expression as is
	return expression, nil
}

// isSingleVariableExpression
func isSingleVariableExpression(expression string) bool {
	return singleVarRegex.MatchString(expression)
}

// evalSingleVariableExpression
func evalSingleVariableExpression(expression string, variables map[string]interface{}) (interface{}, error) {
	parts := singleVarRegex.FindAllStringSubmatch(expression, -1)
	variableName := parts[0][1]
	return variables[variableName], nil
}
