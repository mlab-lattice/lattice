package command

import "fmt"

func NewNoContextSetError() *NoContextSetError {
	return &NoContextSetError{}
}

type NoContextSetError struct{}

func (e *NoContextSetError) Error() string {
	return "no context set"
}

func NewInvalidContextError(context string) *InvalidContextError {
	return &InvalidContextError{context}
}

type InvalidContextError struct {
	Context string
}

func (e *InvalidContextError) Error() string {
	return fmt.Sprintf("invalid context %v", e.Context)
}

func NewContextAlreadyExistsError(context string) *ContextAlreadyExistsError {
	return &ContextAlreadyExistsError{context}
}

type ContextAlreadyExistsError struct {
	Context string
}

func (e *ContextAlreadyExistsError) Error() string {
	return fmt.Sprintf("context %v already exists", e.Context)
}
