package markdown

import (
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/mlab-lattice/lattice/pkg/util/cli"
)

func WriteH1(w io.Writer, header string) {
	fmt.Fprintf(w, "# %s \n", header)
}

func WriteH2(w io.Writer, header string) {
	fmt.Fprintf(w, "## %s \n", header)
}

func WriteH3(w io.Writer, header string) {
	fmt.Fprintf(w, "### %s  \n", header)
}

func WriteFlagArgTableHeader(w io.Writer) {
	fmt.Fprint(w, "| Name | Description | \n")
	fmt.Fprint(w, "| --- | --- | \n")
}

func WriteArgFlagHeader(w io.Writer, text string) {
	fmt.Fprintf(w, "**%s**:  ", text)
}

func WriteEmphasisedText(w io.Writer, text string) {
	fmt.Fprintf(w, "*%s*", text)
}

// WriteFlagTableRow writes flag table row
func WriteFlagTableRow(w io.Writer, flagPtr cli.Flag) {
	currentFlag := flagPtr
	var isBoolFlag bool

	if reflect.TypeOf(currentFlag).String() == "*command.BoolFlag" {
		isBoolFlag = true
	}

	flagRow := "| `"

	flagRow += "--" + currentFlag.GetName()

	if !isBoolFlag {
		flagRow += " " + strings.ToUpper(currentFlag.GetName())
	}

	if currentFlag.GetShort() != "" {
		flagRow += "`, `-" + currentFlag.GetShort()

		if !isBoolFlag {
			flagRow += " " + strings.ToUpper(currentFlag.GetName())
		}
	}

	flagRow += "` "

	if currentFlag.IsRequired() {
		flagRow += "**(required)** "
	}

	flagRow += "| "

	flagRow += currentFlag.GetUsage() + " | \n"

	fmt.Fprint(w, flagRow)
}

// WriteArgTableRow writes arg table row
func WriteArgTableRow(w io.Writer, argPtr cli.Arg) {
	currentArg := argPtr

	argRow := "| `" + currentArg.Name + "` "

	if currentArg.Required {
		argRow += "**(required)** "
	}

	argRow += "| "

	argRow += currentArg.Description + " | \n"

	fmt.Fprint(w, argRow)
}