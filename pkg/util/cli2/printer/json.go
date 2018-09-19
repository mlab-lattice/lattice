package printer

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

func NewJSON(w io.Writer) *JSON {
	return &JSON{writer: w}
}

func NewJSONIndented(w io.Writer, i int) *JSON {
	return &JSON{
		writer: w,
		indent: i,
	}
}

type JSON struct {
	indent int
	writer io.Writer
}

func (j *JSON) Print(v interface{}) error {
	var data []byte
	var err error

	if j.indent == 0 {
		data, err = json.Marshal(v)
	} else {
		data, err = json.MarshalIndent(v, "", strings.Repeat(" ", j.indent))
	}

	if err != nil {
		return err
	}

	_, err = j.writer.Write(data)
	fmt.Fprint(j.writer, "\n")
	return err
}
