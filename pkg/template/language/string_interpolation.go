package language

import (
	"fmt"
	"regexp"
	"strings"
)

// Contains functions needed for evaluating string expressions

// used for matching a single variable reference expression. i.e. ${var}
var varDefRegex = regexp.MustCompile(`\$\{([a-zA-Z_$][a-zA-Z_$.0-9]*)\}`)
var singleVarRegex = regexp.MustCompile(`^\$\{([a-zA-Z_$][a-zA-Z_$.0-9]*)\}$`)

// evalStringExpression
func evalStringExpression(expression string, variables map[string]interface{}) (interface{}, error) {

	if isSingleVariableExpression(expression) {
		return evalSingleVariableExpression(expression, variables)
	}

	// otherwise just return the expression as is
	return replaceAllVariables(expression, variables), nil
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

// replaceAllVariables
func replaceAllVariables(expression string, variables map[string]interface{}) string {
	varDefMatches := varDefRegex.FindAllStringSubmatch(expression, -1)
	result := expression
	for _, group := range varDefMatches {
		variableName := group[1]
		result = replaceVariable(result, variableName, variables)

	}

	return result
}

// replaceVariable
func replaceVariable(expression string, variableName string, variables map[string]interface{}) string {
	varDef := fmt.Sprintf("${%s}", variableName)
	val := getVariableStringValue(variableName, variables)
	return strings.Replace(expression, varDef, val, -1)
}

// getVariableStringValue
func getVariableStringValue(variableName string, variables map[string]interface{}) string {

	if val, exists := variables[variableName]; exists {
		return fmt.Sprintf("%v", val)
	}

	if strings.Contains(variableName, ".") {
		parts := strings.Split(variableName, ".")
		firstVar := parts[0]
		last := strings.Join(parts[1:], ".")

		if newVariables, exists := variables[firstVar]; exists {
			if newVariablesMap, isVarMap := newVariables.(map[string]interface{}); isVarMap {
				return getVariableStringValue(last, newVariablesMap)
			}
		}
	}

	// Unable to determine variable value
	return ""
}
