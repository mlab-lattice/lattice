package printer

import (
	"encoding/json"
	"io"
)

type JSON struct {
	Value interface{}
}

func (j *JSON) Print(writer io.Writer) error {
	data, err := json.MarshalIndent(j.Value, "", "  ")
	if err != nil {
		return err
	}

	_, err = writer.Write(data)
	return err
}
