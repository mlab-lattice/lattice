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
)

type (
	TableAlignment int
	TableRow       []string
)

const (
	TableAlignLeft  = tablewriter.ALIGN_LEFT
	TableAlignRight = tablewriter.ALIGN_RIGHT
)

func NewTable(w io.Writer, columns []TableColumn) *Table {
	return &Table{
		columns: columns,
		writer:  w,
	}
}

type Table struct {
	columns []TableColumn
	rows    []TableRow

	writer     io.Writer
	lastHeight int
}

type TableColumn struct {
	Header    string
	Alignment TableAlignment
}

func (t *Table) AppendRow(row TableRow) {
	t.rows = append(t.rows, row)
}

func (t *Table) AppendRows(rows []TableRow) {
	t.rows = append(t.rows, rows...)
}

func (t *Table) ClearRows() {
	t.rows = []TableRow{}
}

func (t *Table) ReplaceRows(rows []TableRow) {
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

	var headers []string
	var headerColors []tablewriter.Colors
	var alignments []int
	for _, c := range t.columns {
		headers = append(headers, strings.ToUpper(c.Header))
		headerColors = append(headerColors, translateColor(color.Bold))
		//columnColors = append(columnColors, translateColor(c.Color))
		alignments = append(alignments, int(c.Alignment))
	}

	table.SetRowLine(false)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)

	table.SetHeader(headers)
	table.SetAutoFormatHeaders(false)
	table.SetBorder(false)
	table.SetCenterSeparator(color.BlackString(" "))
	table.SetColumnSeparator(color.BlackString(" "))
	table.SetRowSeparator(color.BlackString("-"))
	table.SetAutoWrapText(false)
	table.SetReflowDuringAutoWrap(false)

	table.SetHeaderColor(headerColors...)
	table.SetColumnAlignment(alignments)

	for _, r := range t.rows {
		table.Append(r)
	}

	fmt.Fprintln(w, "")
	table.Render()
	return nil
}

func (t *Table) Rewrite() {
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

func (t *Table) Overwrite(rows []TableRow) {
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
