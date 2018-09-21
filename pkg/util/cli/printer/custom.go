package printer

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/buger/goterm"
)

func NewCustom(w io.Writer) *Custom {
	return &Custom{writer: w}
}

type Custom struct {
	writer     io.Writer
	lastHeight int
}

func (t *Custom) Print(v string) {
	t.print(t.writer, v)
}

func (t *Custom) print(w io.Writer, v string) {
	fmt.Fprintf(w, v)
}

func (t *Custom) Overwrite(v string) {
	// read the new printer's output
	var b bytes.Buffer
	t.print(&b, v)

	// for each line written last time stream was called,
	// return cursor to start of line and clear the rest of the line
	for i := 0; i <= t.lastHeight; i++ {
		if i != 0 {
			goterm.MoveCursorUp(1)
			goterm.ResetLine("")
		}
	}

	// print the output we buffered
	output := b.String()
	goterm.Print(output)
	goterm.Flush()

	t.lastHeight = len(strings.Split(output, "\n"))
}
