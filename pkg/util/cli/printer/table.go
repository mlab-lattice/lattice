package printer

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strings"
	"errors"

	"github.com/buger/goterm"
	"github.com/mattn/go-runewidth"

	"github.com/mlab-lattice/lattice/pkg/util/cli/color"
)

type Table struct {
	Rows      [][]string
	Headers   []string
	nCols     int
	colWidths []int
}

func (t *Table) Print(writer io.Writer) error {
	var w int
	var cellWidth int

	// Number of cols = number of cells in header
	t.nCols = len(t.Headers)
	// Check number of cells in each row match
	if err := t.checkRowLengths(); err != nil {
		return err
	}

	// Apply style to header
	t.formatHeader()

	// Prepend headers to rows for printing
	t.Rows = append([][]string{t.Headers}, t.Rows...)

	// Calculate width of columns from max width of cells
	for i := 0; i < t.nCols; i++ {
		w = 0
		for _, row := range t.Rows {
			cellWidth = runeWidth(row[i])
			if cellWidth > w {
				w = cellWidth
			}
		}
		t.colWidths = append(t.colWidths, w)
	}

	// Add hyphen break after headers
	t.Rows = t.getRowsWithHeaderBreak()

	// Print rows
	for _, row := range t.Rows {
		for col, cell := range row {
			fmt.Fprint(writer, pad(cell, t.colWidths[col])+"   ")
		}
		fmt.Fprint(writer, "\n")
	}

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

func (t *Table) formatHeader() {
	for col, cell := range t.Headers {
		t.Headers[col] = color.Bold(cell)
	}
}

func (t *Table) getRowsWithHeaderBreak() [][]string {
	var headerBreak []string
	for _, w := range t.colWidths {
		headerBreak = append(headerBreak, strings.Repeat("-", w))
	}
	return append(t.Rows[:1], append([][]string{headerBreak}, t.Rows[1:]...)...)
}

func (t *Table) checkRowLengths() error {
	for _, row := range t.Rows {
		fmt.Println(len(row), t.nCols)
		if len(row) != t.nCols {
			return errors.New("Table formatting error: Number of cells do not match. Run with -o json to see unformatted output.")
		}
	}
	return nil
}

// Pad spaces right
func pad(s string, width int) string {
	difference := width - runeWidth(s)
	return s + strings.Repeat(" ", difference)
}

// Regex for ANSI CSI Sequences (specifically SGRs and ELs)
// See https://en.wikipedia.org/wiki/ANSI_escape_code#CSI_sequences
var ansi = regexp.MustCompile("\033\\[(?:[0-9]{1,3}(?:;[0-9]{1,3})*)?[m|K]")

// Returns the rune width of a string. Ignores ANSI SGRs.
// Takes variable width East Asian characters into account.
func runeWidth(s string) int {
	return runewidth.StringWidth(ansi.ReplaceAllLiteralString(s, ""))
}
