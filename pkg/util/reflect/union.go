package reflect

import (
	"fmt"
	"reflect"
	"strings"
)

type InvalidUnionArgumentError struct{}

func (e *InvalidUnionArgumentError) Error() string {
	return fmt.Sprintf("ValidateUnion expects a struct or pointer to a struct")
}

type InvalidUnionFieldError struct {
	Field string
}

func (e *InvalidUnionFieldError) Error() string {
	return fmt.Sprintf("ValidateUnion expects all fields to be pointers, but %v was not", e.Field)
}

type InvalidUnionNoFieldSetError struct{}

func (e *InvalidUnionNoFieldSetError) Error() string {
	return fmt.Sprintf("no fields of the union are set")
}

type InvalidUnionMultipleFieldSetError struct {
	Set []string
}

func (e *InvalidUnionMultipleFieldSetError) Error() string {
	return fmt.Sprintf("multiple fields of the union are set: %s", strings.Join(e.Set, ", "))
}

func ValidateUnion(i interface{}) error {
	// helpful bits: https://gist.github.com/justincase/5469009, (and of course https://golang.org/pkg/reflect)
	v := reflect.ValueOf(i)
	switch v.Kind() {
	case reflect.Struct:
		// nothing to do

	case reflect.Ptr:
		v = v.Elem()

		if v.Kind() != reflect.Struct {
			return &InvalidUnionArgumentError{}
		}

	default:
		return &InvalidUnionArgumentError{}
	}

	t := v.Type()

	var set []string
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		name := t.Field(i).Name
		if f.Kind() != reflect.Ptr {
			return &InvalidUnionFieldError{name}
		}

		if !f.IsNil() {
			set = append(set, name)
		}

	}

	if len(set) == 0 {
		return &InvalidUnionNoFieldSetError{}
	}

	if len(set) > 1 {
		return &InvalidUnionMultipleFieldSetError{Set: set}
	}

	return nil
}
