package printer

import (
	"io"

	"github.com/olekukonko/tablewriter"
)

type Table struct {
	Headers []string
	Rows    [][]string
	table   *tablewriter.Table
}

func (t *Table) Print(writer io.Writer) error {
	table := tablewriter.NewWriter(writer)
	table.SetRowLine(true)
	table.SetAlignment(tablewriter.ALIGN_CENTER)

	table.SetHeader(t.Headers)
	table.SetAutoFormatHeaders(false)
	table.AppendBulk(t.Rows)

	table.Render()
	return nil
}
