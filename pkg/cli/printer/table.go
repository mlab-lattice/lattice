package printer

import (
	"io"
	"fmt"

	"github.com/tfogo/tablewriter"
  "github.com/fatih/color"
)

type Table struct {
	Headers 				[]string
	Rows    				[][]string
	HeaderColors 		[]tablewriter.Colors
	ColumnColors		[]tablewriter.Colors
	ColumnAlignment []int
	table   				*tablewriter.Table
}

func (t *Table) Print(writer io.Writer) error {
	
	FgHiBlack := color.New(color.FgHiBlack).SprintFunc()
	
  table := tablewriter.NewWriter(writer)
	
	table.SetRowLine(false)
	table.SetAlignment(tablewriter.ALIGN_CENTER)
	

	table.SetHeader(t.Headers)
	table.SetAutoFormatHeaders(false)
  table.SetBorder(false)
  table.SetCenterSeparator(FgHiBlack("|"))
  table.SetColumnSeparator(FgHiBlack("|"))
  table.SetRowSeparator(FgHiBlack("-"))
	table.SetAutoWrapText(true)
	table.SetReflowDuringAutoWrap(false)

	table.SetHeaderColor(t.HeaderColors...)
	table.SetColumnColor(t.ColumnColors...)
	table.SetColumnAlignment(t.ColumnAlignment)

	table.AppendBulk(t.Rows)
	
	fmt.Fprintln(writer, "")
	table.Render()
	return nil
}
