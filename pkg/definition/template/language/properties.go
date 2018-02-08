package language

import (
	"strings"

	"github.com/tidwall/gjson"
)

// PropertyMetadata holds information about a certain property
type PropertyMetadata struct {
	propertyPath         string       // the absolute property path
	relativePropertyPath string       // property path relative to the template containing this property
	resource             *urlResource // resource that holds the property being evaluated
}

// PropertyName returns the base property name
func (pm *PropertyMetadata) PropertyName() string {
	props := strings.Split(pm.propertyPath, ".")
	return props[len(props)-1]
}

// PropertyPath absoulte property path
func (pm *PropertyMetadata) PropertyPath() string {
	return pm.propertyPath
}

// RelativePropertyPath relative property path in the template containing the actual property.
func (pm *PropertyMetadata) RelativePropertyPath() string {
	return pm.relativePropertyPath
}

// TemplateURL template url of the template holding this property
func (pm *PropertyMetadata) TemplateURL() string {
	return pm.resource.url
}

// LineNumber line number of the property
func (pm *PropertyMetadata) LineNumber() int {
	return getPropertyPathLineNumber(pm.resource, pm.relativePropertyPath)
}

// getPropertyPathLineNumber utility method for getting a line number of the property
func getPropertyPathLineNumber(resource *urlResource, relativePropertyPath string) int {

	value := gjson.Get(string(resource.bytes), relativePropertyPath)
	return getResultLine(resource.bytes, value)
}

// getResultLine
func getResultLine(bytes []byte, value gjson.Result) int {
	line := 1
	for i := range bytes {
		if i >= value.Index {
			break
		}

		if bytes[i] == '\n' {
			line++
		}
	}

	return line
}

// getParentPropertyPath returns the parent property path of the specified property
func getParentPropertyPath(propertyPath string) string {
	parts := strings.Split(propertyPath, ".")
	if len(parts) == 1 {
		return ""
	}
	return strings.Join(parts[:len(parts)-1], ".")
}
