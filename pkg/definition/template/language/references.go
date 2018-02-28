package language

import (
	"fmt"
)

const (
	referenceKey          = "__reference"
	templateReferencesKey = "__references"
)

type Reference map[string]interface{}

func (r Reference) getTarget() string {
	return r[referenceKey].(string)
}

func newReferenceObject(reference string) Reference {
	return Reference{
		referenceKey: reference,
	}
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

	if reference, isReference := value.(Reference); isReference && isReferenceTargetInTemplate(reference, template, env) {
		recipient := propertyPath

		return []interface{}{
			newReferenceEntry(reference.getTarget(), recipient),
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

// isReferenceTargetInTemplate
func isReferenceTargetInTemplate(reference Reference, template *Template, env *environment) bool {
	meta := env.getPropertyMetaData(reference.getTarget())
	return meta != nil && meta.template != nil && meta.template == template
}
