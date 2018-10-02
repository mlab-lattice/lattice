package command

import "fmt"

func NewNoContextSetError() *NoContextSetError {
	return &NoContextSetError{}
}

// NoContextSetError indicates that there is no Context set in the Configuration.
type NoContextSetError struct{}

func (e *NoContextSetError) Error() string {
	return "no context set"
}

func NewInvalidContextError(context string) *InvalidContextError {
	return &InvalidContextError{context}
}

// InvalidContextError indicates that the supplied Context does not exist.
type InvalidContextError struct {
	Context string
}

func (e *InvalidContextError) Error() string {
	return fmt.Sprintf("invalid context %v", e.Context)
}

func NewContextAlreadyExistsError(context string) *ContextAlreadyExistsError {
	return &ContextAlreadyExistsError{context}
}

// ContextAlreadyExistsError indicates that the supplied Context already exist.
type ContextAlreadyExistsError struct {
	Context string
}

func (e *ContextAlreadyExistsError) Error() string {
	return fmt.Sprintf("context %v already exists", e.Context)
}
