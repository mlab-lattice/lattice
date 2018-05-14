package printer

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"regexp"
	//"text/tabwriter"
	//"unicode/utf8"
	//"time"
	//"sync"

	"github.com/buger/goterm"
	//"github.com/fatih/color"
	"github.com/tfogo/tablewriter"
	"github.com/mattn/go-runewidth"

	"github.com/mlab-lattice/lattice/pkg/util/cli/color"

	//"k8s.io/kubernetes/pkg/printers"
)

type Table struct {
	Headers         []string
	Rows            [][]string
	HeaderColors    []tablewriter.Colors
	ColumnColors    []tablewriter.Colors
	ColumnAlignment []int
}

// func (t *Table) Print(writer io.Writer) error {
//
// 	FgHiBlack := color.New(color.FgHiBlack).SprintFunc()
//
// 	table := tablewriter.NewWriter(writer)
//
// 	var hs []string
// 	for _, h := range t.Headers {
// 		hs = append(hs, strings.ToUpper(h))
// 	}
//
// 	t.Headers = hs
//
// 	table.SetRowLine(false)
// 	table.SetAlignment(tablewriter.ALIGN_LEFT)
// 	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
//
// 	table.SetHeader(t.Headers)
// 	table.SetAutoFormatHeaders(false)
// 	table.SetBorder(false)
// 	//table.SetHeaderLine(false)
// 	table.SetCenterSeparator(FgHiBlack(" "))
// 	table.SetColumnSeparator(FgHiBlack(" "))
// 	table.SetRowSeparator(FgHiBlack("-"))
// 	table.SetAutoWrapText(false)
// 	table.SetReflowDuringAutoWrap(false)
//
// 	table.SetHeaderColor(t.HeaderColors...)
// 	table.SetColumnColor(t.ColumnColors...)
// 	table.SetColumnAlignment(t.ColumnAlignment)
//
// 	table.AppendBulk(t.Rows)
//
// 	fmt.Fprintln(writer, "")
// 	table.Render()
// 	return nil
// }

func (t *Table) Print(writer io.Writer) error {
	// w := NewWriter(writer, 0, 0, 3, ' ', 0)
	//
	// fmt.Println(t.Headers)
	// headers := strings.Join(t.Headers, "\t")
	//
	//
	// fmt.Fprintln(w, headers)
	//
	// for _, row := range t.Rows {
	// 	rowString := strings.Join(row, "\t")
	// 	fmt.Fprintln(w, rowString)
	// }
	//
	//
	// w.Flush()

	var tab2 Table2
	var allRows [][]string
	allRows = append(allRows, t.Headers)
	allRows = append(allRows, t.Rows...)
	tab2.rows = allRows
	tab2.nCols = len(t.Rows[0])
	tab2.columnColors = []color.Color{color.ID}
	tab2.Print(writer)

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
			// Return cursor to start of line and clear the rest of the line
			// Waiting on burger/goterm#23 to be merged to use ResetLine
			goterm.Print("\r\033[K")
		}
	}

	goterm.Print(output)
	goterm.Flush() // TODO: Fix for large outputs (e.g. systems:builds)

	return len(strings.Split(output, "\n"))
}





type Table2 struct {
	rows [][]string
	nCols int
	colWidths []int
	columnColors []color.Color
}

func (tab Table2) Print(writer io.Writer) {


	var w int
	var cellWidth int

	tab.formatHeader()
	tab.formatRows()




	for i := 0; i < tab.nCols; i++ {
		w = 0
		for _, row := range tab.rows {
			cellWidth = runeWidth(row[i])
			if cellWidth > w {
				w = cellWidth
			}
		}
		tab.colWidths = append(tab.colWidths, w)
	}

	tab.rows = tab.addHeaderBreak()

	for _, row := range tab.rows {
		for col, cell := range row {
			fmt.Fprint(writer, pad(cell, tab.colWidths[col]) + "   ")
		}
		fmt.Fprint(writer, "\n")
	}
}

func (tab Table2) formatHeader() {
	for col, cell := range tab.rows[0] {
		tab.rows[0][col] = color.Bold(cell)
	}
}

func (tab Table2) formatRows() {
	for n, _ := range tab.rows[1:] {
		tab.rows[n+1][0] = tab.columnColors[0](tab.rows[n+1][0])
		// for col, cell := range row {
		// 	tab.rows[n][col] = tab.columnColors[col](cell)
		// }
	}
}

func (tab Table2) addHeaderBreak() [][]string {
	var headerBreak []string
	for _, w := range tab.colWidths {
		headerBreak = append(headerBreak, strings.Repeat("-", w))
	}
	return append(tab.rows[:1], append([][]string{headerBreak}, tab.rows[1:]...)...)
}

func pad(s string, width int) string {
	difference := width - runeWidth(s)
	return s + strings.Repeat(" ", difference)
}

// Regex for control sequences
var ansi = regexp.MustCompile("\033\\[(?:[0-9]{1,3}(?:;[0-9]{1,3})*)?[m|K]")

func runeWidth(s string) int {
	return runewidth.StringWidth(ansi.ReplaceAllLiteralString(s, ""))
}
