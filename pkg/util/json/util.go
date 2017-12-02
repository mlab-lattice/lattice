package json

type FieldBytes struct {
	Name      string
	Bytes     []byte
	OmitEmpty bool
}

func GenerateObjectBytes(fields []FieldBytes) []byte {
	b := []byte(`{`)

	first := true
	for _, field := range fields {
		if field.Bytes == nil && field.OmitEmpty {
			continue
		}

		if first {
			first = false
		} else {
			b = append(b, []byte(`,`)...)
		}

		b = append(b, []byte(`"`)...)
		b = append(b, []byte(field.Name)...)
		b = append(b, []byte(`":`)...)
		b = append(b, []byte(field.Bytes)...)
	}

	b = append(b, []byte(`}`)...)
	return b
}

func GenerateArrayBytes(objects [][]byte) []byte {
	b := []byte(`[`)

	first := true
	for _, object := range objects {
		if first {
			first = false
		} else {
			b = append(b, []byte(`,`)...)
		}

		b = append(b, []byte(object)...)
	}

	b = append(b, []byte(`]`)...)
	return b
}
