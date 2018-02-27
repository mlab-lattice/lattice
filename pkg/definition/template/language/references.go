package language

import (
	"fmt"
)

const (
	referenceKey          = "__reference"
	templateReferencesKey = "__references"
)

// isReferenceObject
func isReferenceObject(val interface{}) bool {
	if val == nil {
		return false
	}

	if mapVal, isMap := val.(map[string]interface{}); isMap {
		return len(mapVal) == 1 && mapVal[referenceKey] != nil
	}

	return false
}

// getReferenceObjectTarget
func getReferenceObjectTarget(val interface{}) string {
	mapVal := val.(map[string]interface{})
	return mapVal[referenceKey].(string)
}

// newReferenceEntry
func newReferenceEntry(target string, recipient string) map[string]interface{} {
	return map[string]interface{}{
		"target":    target,
		"recipient": recipient,
	}
}

// findReferencesInTemplate
func findReferencesInTemplate(template *Template, value interface{}, env *environment) ([]interface{}, error) {

	return findReferences(template, value, env.getCurrentPropertyPath(), env)
}

// findReferences
func findReferences(template *Template, value interface{}, propertyPath string, env *environment) ([]interface{}, error) {

	if isReferenceObject(value) && isReferenceTargetInTemplate(value, template, env) {

		target := getReferenceObjectTarget(value)
		recipient := propertyPath

		err := validateReference(target, recipient, template, env)

		if err != nil {
			return nil, err
		}

		return []interface{}{
			newReferenceEntry(target, recipient),
		}, nil
	}

	if valMap, isMap := value.(map[string]interface{}); isMap { // Maps
		return findReferencesInMap(template, valMap, propertyPath, env)

	} else if valArr, isArray := value.([]interface{}); isArray { // Arrays
		return findReferencesInArray(template, valArr, propertyPath, env)

	}

	// default, return empty array
	return make([]interface{}, 0), nil
}

// findReferencesInMap
func findReferencesInMap(template *Template, m map[string]interface{}, propertyPath string, env *environment) ([]interface{}, error) {

	references := make([]interface{}, 0)
	for k, v := range m {
		var currentPropertyPath string
		if propertyPath != "" {
			currentPropertyPath = fmt.Sprintf("%v.%v", propertyPath, k)
		} else {
			currentPropertyPath = k
		}

		childRefs, err := findReferences(template, v, currentPropertyPath, env)
		if err != nil {
			return nil, err
		}
		references = append(references, childRefs...)

	}

	return references, nil
}

// findReferencesInArray
func findReferencesInArray(template *Template, arr []interface{}, propertyPath string, env *environment) ([]interface{}, error) {
	references := make([]interface{}, 0)
	for i, item := range arr {
		itemPropPath := fmt.Sprintf("%v.%v", propertyPath, i)
		childRefs, err := findReferences(template, item, itemPropPath, env)

		if err != nil {
			return nil, err
		}

		references = append(references, childRefs...)

	}

	return references, nil
}

// validateReference
func validateReference(target string, recipient string, template *Template, env *environment) error {
	recipientMeta := env.getPropertyMetaData(recipient)
	if recipientMeta.template == template {
		// references in the same template are ok. nothing to do here
		return nil
	}

	// TODO validate
	return nil

}

// isReferenceTargetInTemplate
func isReferenceTargetInTemplate(reference interface{}, template *Template, env *environment) bool {
	target := getReferenceObjectTarget(reference)
	meta := env.getPropertyMetaData(target)
	return meta != nil && meta.template != nil && meta.template == template
}
