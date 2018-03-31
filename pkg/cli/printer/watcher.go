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

type Watcher2 interface {
	Watch(chan Interface, io.Writer, chan bool)
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

type ScrollingWatcher2 struct {
}

func (w *ScrollingWatcher2) Watch(printers chan Interface, writer io.Writer, printerRenderedChan chan bool) {
	for printer := range printers {
		printer.Print(writer)
		writer.Write([]byte("\n"))
		printerRenderedChan <- true
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
				
				goterm.ResetLine("")
			}
			goterm.MoveCursorUp(1)
		}

		goterm.Print(output)
		goterm.Flush()

		//lastHeight = len(strings.Split(output, "\n"))
		lastHeight = strings.Count(output, "\n")
	}
}

type OverwrittingTerminalWatcher2 struct {
}

func (w *OverwrittingTerminalWatcher2) Watch(printers chan Interface, writer io.Writer, printerRenderedChan chan bool) {
	lastHeight := 0
	var b bytes.Buffer
	// firstPrint := true

	for printer := range printers {
		// Read the new printer's output
		printer.Print(&b)
		output := b.String()

		// Remove the new printer's output from the buffer
		b.Truncate(0)

		// Clear the previous render's
		//goterm.MoveCursorUp(1)
		
		for i := 0; i <= lastHeight; i++ {
			if i != 0 {
				goterm.MoveCursorUp(1)
				goterm.ResetLine("")
			}
		}

		goterm.Print(output)
		//goterm.MoveCursorDown(1)
		goterm.Flush()
		
		printerRenderedChan <- true
		
		// if firstPrint {
		// 	printCompleteChan <- "first"
		// } else {
		// 	printCompleteChan <- "subsequent"
		// }
		
		// firstPrint = false
		lastHeight = len(strings.Split(output, "\n"))
	}
}
