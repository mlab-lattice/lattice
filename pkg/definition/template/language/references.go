package language

import "github.com/mlab-lattice/system/bazel-system/external/go_sdk/src/fmt"

func isReferenceObject(val interface{}) bool {
	if val == nil {
		return false
	}

	if mapVal, isMap := val.(map[string]interface{}); isMap {
		return len(mapVal) == 1 && mapVal["reference"] != nil
	}

	return false
}

func getReferenceObjectTarget(val interface{}) string {
	mapVal := val.(map[string]interface{})
	return mapVal["reference"].(string)
}

func newReferenceEntry(target string, recipient string) map[string]interface{} {
	return map[string]interface{}{
		"target":    target,
		"recipient": recipient,
	}
}
func findReferences(url string, o interface{}, propertyPath string, env *environment) ([]interface{}, error) {

	propertyMeta := env.getPropertyMetaData(propertyPath)

	if isReferenceObject(o) && propertyMeta != nil && propertyMeta.resource != nil && propertyMeta.resource.url == url {

		err := validateReference(url, o, propertyPath, env)

		if err != nil {
			return nil, err
		}
		target := getReferenceObjectTarget(o)
		return []interface{}{
			newReferenceEntry(target, propertyPath),
		}, nil
	}

	if valMap, isMap := o.(map[string]interface{}); isMap { // Maps
		return findReferencesInMap(url, valMap, propertyPath, env)

	} else if valArr, isArray := o.([]interface{}); isArray { // Arrays
		return findReferencesInArray(url, valArr, propertyPath, env)

	}

	// default, return empty array
	return make([]interface{}, 0), nil
}

func findReferencesInMap(url string, m map[string]interface{}, propertyPath string, env *environment) ([]interface{}, error) {

	references := make([]interface{}, 0)
	for k, v := range m {
		var currentPropertyPath string
		if propertyPath != "" {
			currentPropertyPath = fmt.Sprintf("%v.%v", propertyPath, k)
		} else {
			currentPropertyPath = k
		}

		childRefs, err := findReferences(url, v, currentPropertyPath, env)
		if err != nil {
			return nil, err
		}
		references = append(references, childRefs...)

	}

	return references, nil
}

func findReferencesInArray(url string, arr []interface{}, propertyPath string, env *environment) ([]interface{}, error) {
	references := make([]interface{}, 0)
	for i, item := range arr {
		itemPropPath := fmt.Sprintf("%v.%v", propertyPath, i)
		childRefs, err := findReferences(url, item, itemPropPath, env)

		if err != nil {
			return nil, err
		}

		references = append(references, childRefs...)

	}

	return references, nil
}

func validateReference(url string, o interface{}, propertyPath string, env *environment) error {
	return nil
}
