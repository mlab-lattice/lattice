package printer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
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

// TODO: Refactor this part of the interface, it's currently ugly
// Not overwriting, we just print json objects on new lines
func (j *JSON) Overwrite(b bytes.Buffer, lastHeight int) (error, int) {
	if err := j.Print(os.Stdout); err != nil {
		return err, 0
	}
	fmt.Print("\n")
	return nil, 1 // Not used in JSON
}
