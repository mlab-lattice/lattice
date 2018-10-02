package printer

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	//"time"
	//"sync"

	"github.com/mlab-lattice/lattice/pkg/util/cli/color"

	"github.com/buger/goterm"
	"github.com/olekukonko/tablewriter"
	"os"
)

const (
	TableAlignLeft  = tablewriter.ALIGN_LEFT
	TableAlignRight = tablewriter.ALIGN_RIGHT
)

func NewTable(w io.Writer, columns []string) *Table {
	return &Table{
		columns: columns,
		writer:  w,
	}
}

type Table struct {
	columns []string
	rows    [][]string

	writer     io.Writer
	lastHeight int
}

func (t *Table) AppendRow(row []string) {
	t.rows = append(t.rows, row)
}

func (t *Table) AppendRows(rows [][]string) {
	t.rows = append(t.rows, rows...)
}

func (t *Table) ClearRows() {
	t.rows = [][]string{}
}

func (t *Table) ReplaceRows(rows [][]string) {
	t.ClearRows()
	t.AppendRows(rows)
}

func (t *Table) Print() error {
	return t.print(t.writer)
}

func (t *Table) print(w io.Writer) error {
	// right now we're creating a new table on each write
	// this probably isn't necessary but for now need to do
	// it so we can write the table to a buffer for Rewrite
	table := tablewriter.NewWriter(w)

	var headerColors []tablewriter.Colors
	for range t.columns {
		headerColors = append(headerColors, translateColor(color.Bold))
	}

	table.SetRowLine(false)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)

	table.SetHeader(t.columns)
	table.SetAutoFormatHeaders(false)
	table.SetBorder(false)
	table.SetCenterSeparator(color.BlackString(" "))
	table.SetColumnSeparator(color.BlackString(" "))
	table.SetRowSeparator(color.BlackString("-"))
	table.SetAutoWrapText(false)
	table.SetReflowDuringAutoWrap(false)

	table.SetHeaderColor(headerColors...)

	for _, r := range t.rows {
		table.Append(r)
	}

	fmt.Fprintln(w, "")
	table.Render()
	return nil
}

func (t *Table) Rewrite() {
	if t.writer != os.Stdout {
		panic("cannot call Rewrite on a writer that is not stdout")
	}

	// read the new printer's output
	var b bytes.Buffer
	t.print(&b)

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

func (t *Table) Overwrite(rows [][]string) {
	t.ReplaceRows(rows)
	t.Rewrite()
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
