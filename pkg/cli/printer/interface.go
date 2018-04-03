package printer

import (
	"io"
	"bytes"
)

type Interface interface {
	Print(writer io.Writer) error
	Overwrite(b bytes.Buffer, lastHeight int) int
}
