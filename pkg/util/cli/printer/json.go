package printer

import (
	"encoding/json"
	"io"
)

type JSON struct {
	Value  interface{}
	Indent int
}

func (j *JSON) Print(writer io.Writer) error {
	var data []byte
	var err error

	if j.Indent == 0 {
		data, err = json.Marshal(j.Value)
	} else {
		data, err = json.MarshalIndent(j.Value, "", "  ")
	}

	if err != nil {
		return err
	}

	_, err = writer.Write(data)
	return err
}
