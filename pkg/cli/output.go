package cli

import "fmt"

type OutputFormat string

const (
	OutputFormatJSON  = "json"
	OutputFormatTable = "table"
)

func newOutputFormatError(format OutputFormat) error {
	return fmt.Errorf("invalid output format: %v", format)
}
