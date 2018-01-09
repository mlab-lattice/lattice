package cli

import "fmt"

type OutputFormat string

const (
	OutputFormatJSON  = "json"
	OutputFormatTable = "table"
)

func NewOutputFormatError(format OutputFormat) error {
	return fmt.Errorf("invalid output format: %v", format)
}
