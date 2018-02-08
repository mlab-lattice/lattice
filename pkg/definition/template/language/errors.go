package language

import (
	"bytes"
	"fmt"
)

// PropertyEvalError returned when there is an evaluation error of a property in a map or array element
type PropertyEvalError struct {
	PropertyMetaData *PropertyMetadata
	err              error
}

// Error() implement the error interface
func (e *PropertyEvalError) Error() string {
	var msgBuf bytes.Buffer

	msgBuf.WriteString(fmt.Sprintf("Evaluation error for '%v': %v.", e.PropertyMetaData.PropertyName(), e.err))

	msgBuf.WriteString(fmt.Sprintf(" File: '%v'", e.PropertyMetaData.TemplateURL()))
	msgBuf.WriteString(fmt.Sprintf(" at line %v", e.PropertyMetaData.LineNumber()))

	return msgBuf.String()

}

// wrapWithPropertyEvalError wraps the current error with a PropertyEvalError UNLESS err is already a PropertyEvalError
func wrapWithPropertyEvalError(err error, propertyPath string, env *environment) error {
	if err == nil {
		return nil
	}

	if _, isEvalError := err.(*PropertyEvalError); isEvalError {
		return err
	}

	return newPropertyEvalError(err, propertyPath, env)
}

// newPropertyEvalError creates a new PropertyEvalError
func newPropertyEvalError(err error, propertyPath string, env *environment) error {
	if err == nil {
		return nil
	}

	return &PropertyEvalError{
		err:              err,
		PropertyMetaData: env.getPropertyMetaData(propertyPath),
	}
}
