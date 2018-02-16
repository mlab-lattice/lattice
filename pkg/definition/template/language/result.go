package language

// Result evaluation result
type Result struct {
	value interface{}
	env   *environment
}

// Value returns the raw evaluation value
func (r *Result) Value() interface{} {
	return r.value
}

// ValueAsMap returns the eval value as a map.
func (r *Result) ValueAsMap() map[string]interface{} {
	return r.value.(map[string]interface{})
}

// GetPropertyMetadata returns metadata information about a specific property path.
func (r *Result) GetPropertyMetadata(propertyPath string) *PropertyMetadata {
	return r.env.getPropertyMetaData(propertyPath)
}

// newResult returns a new Result struct
func newResult(value interface{}, env *environment) *Result {
	return &Result{
		value: value,
		env:   env,
	}
}
