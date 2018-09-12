package printer

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	//"time"
	//"sync"

	"github.com/mlab-lattice/lattice/pkg/util/cli2/color"

	"github.com/buger/goterm"
	"github.com/tfogo/tablewriter"
)

type (
	TableAlignment int
)

const (
	TableAlignLeft  = tablewriter.ALIGN_LEFT
	TableAlignRight = tablewriter.ALIGN_RIGHT
)

type Table struct {
	Columns []TableColumn
	rows    [][]string

	lastHeight int
}

type TableColumn struct {
	Header    string
	Color     color.Color
	Alignment TableAlignment
}

func (t *Table) AppendRow(row []string) {
	t.rows = append(t.rows, row)
}

func (t *Table) AppendRows(rows [][]string) {
	t.rows = append(t.rows, rows...)
}

func (t *Table) Print(writer io.Writer) error {
	table := tablewriter.NewWriter(writer)

	var headers []string
	var headerColors []tablewriter.Colors
	var columnColors []tablewriter.Colors
	var alignments []int
	for _, c := range t.Columns {
		headers = append(headers, strings.ToUpper(c.Header))
		headerColors = append(headerColors, translateColor(color.Bold))
		columnColors = append(columnColors, translateColor(c.Color))
		alignments = append(alignments, int(c.Alignment))
	}

	table.SetRowLine(false)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)

	table.SetHeader(headers)
	table.SetAutoFormatHeaders(false)
	table.SetBorder(false)
	table.SetCenterSeparator(color.Black(" "))
	table.SetColumnSeparator(color.Black(" "))
	table.SetRowSeparator(color.Black("-"))
	table.SetAutoWrapText(false)
	table.SetReflowDuringAutoWrap(false)

	table.SetHeaderColor(headerColors...)
	table.SetColumnColor(columnColors...)
	table.SetColumnAlignment(alignments)

	table.AppendBulk(t.rows)

	fmt.Fprintln(writer, "")
	table.Render()
	return nil
}

func (t *Table) Stream(w io.Writer) {
	// read the new printer's output
	var b bytes.Buffer
	t.Print(&b)

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
	goterm.Flush() // TODO: Fix for large outputs (e.g. systems:builds)

	t.lastHeight = len(strings.Split(output, "\n"))
}

func translateColor(c color.Color) tablewriter.Colors {
	switch c {
	case color.Success:
		return tablewriter.Color(tablewriter.FgGreenColor)

	case color.BoldHiSuccess:
		return tablewriter.Color(tablewriter.Bold, tablewriter.FgHiGreenColor)

	case color.Failure:
		return tablewriter.Color(tablewriter.FgRedColor)

	case color.BoldHiFailure:
		return tablewriter.Color(tablewriter.Bold, tablewriter.FgHiRedColor)

	case color.Warning:
		return tablewriter.Color(tablewriter.FgYellowColor)

	case color.BoldHiWarning:
		return tablewriter.Color(tablewriter.Bold, tablewriter.FgHiYellowColor)

	case color.ID:
		return tablewriter.Color(tablewriter.FgHiCyanColor)

	case color.Bold:
		return tablewriter.Color(tablewriter.Bold)

	default:
		return tablewriter.Color()
	}
}
