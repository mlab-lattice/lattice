package reflect

import (
	"reflect"
	"testing"
)

func TestValidateUnion(t *testing.T) {
	// test basic validation
	type foo struct {
		Bar  *int
		Baz  *string
		buzz *bool
	}

	f := &foo{}
	err := ValidateUnion(f)
	if err == nil {
		t.Fatalf("expected to get an error when passing struct with no fields set")
	}
	_, ok := err.(*InvalidUnionNoFieldSetError)
	if !ok {
		t.Fatalf("expected to get an InvalidUnionNoFieldSetError when passing struct with no fields set but got %v", err)
	}

	one := 1
	f = &foo{
		Bar: &one,
	}

	// test pointer to struct
	err = ValidateUnion(f)
	if err != nil {
		t.Fatalf("expected to get no error for Bar being set only but got %v", err)
	}

	// test struct by value
	err = ValidateUnion(*f)
	if err != nil {
		t.Fatalf("expected to get no error for Bar being set only but got %v", err)
	}

	// test pointer to pointer to struct
	err = ValidateUnion(&f)
	if err == nil {
		t.Fatalf("expected to get an error when passing in pointer to pointer to struct")
	}
	_, ok = err.(*InvalidUnionArgumentError)
	if !ok {
		t.Fatalf("expected to get an InvalidUnionArgumentError when passing in pointer to pointer to struct but got %v", err)
	}

	hello := "hello"
	f = &foo{
		Baz: &hello,
	}

	err = ValidateUnion(f)
	if err != nil {
		t.Fatalf("expected to get no error for Baz being set only but got %v", err)
	}

	true := true
	f = &foo{
		buzz: &true,
	}

	err = ValidateUnion(f)
	if err != nil {
		t.Fatalf("expected to get no error for buzz being set only but got %v", err)
	}

	// test multiple fields set
	f = &foo{
		Bar: &one,
		Baz: &hello,
	}
	err = ValidateUnion(f)
	if err == nil {
		t.Fatalf("expected to get an error when passing in pointer to pointer to struct")
	}
	e, ok := err.(*InvalidUnionMultipleFieldSetError)
	if !ok {
		t.Fatalf("expected to get an InvalidUnionArgumentError when passing in pointer to pointer to struct but got %v", err)
	}
	if !reflect.DeepEqual(e.Set, []string{"Bar", "Baz"}) {
		t.Fatalf("unexpected set list: %v", e.Set)
	}

	f = &foo{
		Bar:  &one,
		Baz:  &hello,
		buzz: &true,
	}
	err = ValidateUnion(f)
	if err == nil {
		t.Fatalf("expected to get an error when passing in pointer to pointer to struct")
	}
	e, ok = err.(*InvalidUnionMultipleFieldSetError)
	if !ok {
		t.Fatalf("expected to get an InvalidUnionArgumentError when passing in pointer to pointer to struct but got %v", err)
	}
	if !reflect.DeepEqual(e.Set, []string{"Bar", "Baz", "buzz"}) {
		t.Fatalf("unexpected set list: %v", e.Set)
	}

	// test invalid union struct
	type bar struct {
		Bar  *int
		Baz  *string
		Buzz *bool
		Qux  int
	}
	b := &bar{
		Bar: &one,
	}
	err = ValidateUnion(b)
	if err == nil {
		t.Fatalf("expected to get an error when passing in pointer to pointer to struct")
	}
	_, ok = err.(*InvalidUnionFieldError)
	if !ok {
		t.Fatalf("expected to get an InvalidUnionFieldError when passing in struct with non-pointer field but got %v", err)
	}
}
