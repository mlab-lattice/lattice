package printer

import (
	"io"
	"fmt"
	"bytes"
	"strings"
	//"time"
	//"sync"

	"github.com/tfogo/tablewriter"
  "github.com/fatih/color"
	"github.com/buger/goterm"
)

type Table struct {
	Headers 				[]string
	Rows    				[][]string
	HeaderColors 		[]tablewriter.Colors
	ColumnColors		[]tablewriter.Colors
	ColumnAlignment []int
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
	table.SetAutoWrapText(false)
	table.SetReflowDuringAutoWrap(false)

	table.SetHeaderColor(t.HeaderColors...)
	table.SetColumnColor(t.ColumnColors...)
	table.SetColumnAlignment(t.ColumnAlignment)

	table.AppendBulk(t.Rows)
	
	fmt.Fprintln(writer, "")
	table.Render()
	return nil
}

func (t *Table) Overwrite(b bytes.Buffer, lastHeight int) int {

	// Read the new printer's output
	t.Print(&b)
	output := b.String()
	// Remove the new printer's output from the buffer
	b.Truncate(0)

	for i := 0; i <= lastHeight; i++ {
		if i != 0 {
			goterm.MoveCursorUp(1)
			goterm.ResetLine("")
		}
	}

	goterm.Print(output)
	goterm.Flush() // TODO: Fix for large outputs (e.g. systems:builds)

	return len(strings.Split(output, "\n"))
}
