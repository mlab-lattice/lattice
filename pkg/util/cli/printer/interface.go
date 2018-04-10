package printer

import (
	"io"
)

type Interface interface {
	Print(writer io.Writer) error
}
