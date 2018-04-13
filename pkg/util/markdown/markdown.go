package markdown

import (
	"fmt"
	"io"
)

func WrapH1(header string) string {
	return fmt.Sprintf("# %s", header)
}

func WrapH2(header string) string {
	return fmt.Sprintf("## %s", header)
}

func WrapH3(header string) string {
	return fmt.Sprintf("### %s", header)
}

func WriteTableHeader(w io.Writer, headers []string) {
	for _, header := range headers {
		fmt.Fprintf(w, "| %v ", header)
	}
	fmt.Fprintln(w, "|")

	for range headers {
		fmt.Fprint(w, "| --- ")
	}

	fmt.Fprintln(w, "|")
}

func WriteTableRow(w io.Writer, cells []string) {
	for _, cell := range cells {
		fmt.Fprintf(w, "| %v ", cell)
	}

	fmt.Fprintln(w, "|")
}

func WrapBold(text string) string {
	return fmt.Sprintf("**%s**", text)
}

func WrapItalic(text string) string {
	return fmt.Sprintf("*%s*", text)
}

func WrapInlineCode(text string) string {
	return fmt.Sprintf("`%s`", text)
}
