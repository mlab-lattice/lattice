package printer

import (
	"bytes"
	"io"
	"strings"

	"github.com/buger/goterm"
)

type Watcher interface {
	Watch(chan Interface, io.Writer)
}

// ScrollingWatcher is a watcher that writes the printer's output to
// the writer, followed by a newline.
type ScrollingWatcher struct {
}

func (w *ScrollingWatcher) Watch(printers chan Interface, writer io.Writer) {
	for printer := range printers {
		printer.Print(writer)
		writer.Write([]byte("\n"))
	}
}

// OverwrittingTerminalWatcher is a watcher that flushes the printer's output
// to the terminal, ignoring the writer that is passed in.
// When it receives a new printer, it clears the screen of the previous printer's
// output and prints the new printer's output.
type OverwrittingTerminalWatcher struct {
}

func (w *OverwrittingTerminalWatcher) Watch(printers chan Interface, writer io.Writer) {
	lastHeight := 0
	var b bytes.Buffer

	for printer := range printers {
		// Read the new printer's output
		printer.Print(&b)
		output := b.String()

		// Remove the new printer's output from the buffer
		b.Truncate(0)

		// Clear the previous render's
		for i := 0; i <= lastHeight; i++ {
			if i != 0 {
				goterm.MoveCursorUp(1)
			}
			goterm.ResetLine("")
		}

		goterm.Print(output)
		goterm.Flush()

		lastHeight = len(strings.Split(output, "\n"))
	}
}
