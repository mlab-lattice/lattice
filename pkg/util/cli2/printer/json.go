package printer

import (
	"encoding/json"
	"fmt"
	"io"
)

type JSON struct {
	Value  interface{}
	Indent int
}

func (j *JSON) Print(w io.Writer) error {
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

	_, err = w.Write(data)
	return err
}

// Not overwriting, we just print json objects on new lines
func (j *JSON) Stream(w io.Writer) {
	j.Print(w)
	fmt.Fprint(w, "\n")
}
