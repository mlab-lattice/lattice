package docgen

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/mlab-lattice/lattice/pkg/util/cli"
	"github.com/mlab-lattice/lattice/pkg/util/markdown"
)

var InputDocsDir string

// extra markdown file name
const (
	descriptionFile = "description.md"
	examplesFile    = "examples.md"
)

func GenerateMarkdown(cmd *cli.Command) (io.Reader, error) {
	var buffer bytes.Buffer
	writer := bufio.NewWriter(&buffer)

	err := writeDoc(cmd, writer)
	if err != nil {
		return nil, err
	}

	writer.Flush()

	return bytes.NewReader(buffer.Bytes()), nil
}

// writeDoc writes command tree to file
func writeDoc(bc *cli.Command, writer io.Writer) error {
	// header
	fmt.Fprintf(writer, "%s \n", markdown.WrapH1(bc.Name))

	fmt.Fprintf(writer, "%s \n", markdown.WrapH2("Introduction"))

	// extra description in the intro section
	introMdFileContent, err := getMarkdownFileContent("", descriptionFile)
	if err != nil {
		return err
	}
	if introMdFileContent != "" {
		fmt.Fprintf(writer, "%s \n\n", introMdFileContent)
	}

	fmt.Fprintf(writer, "%s \n", markdown.WrapH2("Commands"))

	// mapping between command name (e.g. `laasctl systems builds`) and the command.
	// the output to file can therefore be sorted by command name (ascending),
	// while still using depth-first tree traversal
	commandMapping := make(map[string]*cli.Command)

	// reads the tree and appends commands into the commandMapping
	for _, childCmd := range bc.Subcommands {
		var cmds []cli.Command
		recurse(childCmd, cmds, writer, commandMapping)
	}

	// uses sorted map keys to iterate over the map in alphabetical order
	keys := make([]string, 0)
	for k := range commandMapping {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, key := range keys {
		err := printCommand(key, commandMapping[key], writer)
		if err != nil {
			return err
		}
	}

	return nil
}

// recurse recursively works through the command tree
// ancestorCommands are commands that precede the current cmd in the hierarchy
func recurse(cmd *cli.Command, ancestorCommands []cli.Command, writer io.Writer, commandMapping map[string]*cli.Command) {
	if cmd.Name == "" {
		return
	}

	// joins consecutive ancestor command names
	var ancestorCmdsStr string
	for _, tempCmd := range ancestorCommands {
		ancestorCmdsStr += tempCmd.Name + ":"
	}
	ancestorCmdsStr += cmd.Name

	commandMapping[ancestorCmdsStr] = cmd

	// adds this command to the list of ancestor commands to be passed to its children when recursing
	ancestorCommands = append(ancestorCommands, *cmd)

	for _, tempCmd := range cmd.Subcommands {
		recurse(tempCmd, ancestorCommands, writer, commandMapping)
	}
}

// printCommand prints the command docs
// fullCmdName includes all command ancestor except the root command (e.g. 'latticectl')
func printCommand(fullCmdName string, cmd *cli.Command, writer io.Writer) error {
	fmt.Fprintf(writer, "%s  \n", markdown.WrapH2(fullCmdName))

	if cmd.Short != "" {
		fmt.Fprintf(writer, "%s  \n\n", markdown.WrapItalic(cmd.Short))
	}

	// includes any extra markdown command description
	descMdFile, err := getMarkdownFileContent(fullCmdName, descriptionFile)
	if err != nil {
		return err
	}

	if descMdFile != "" {
		fmt.Fprintf(writer, "%s \n\n", descMdFile)
	}

	if len(cmd.Args) > 0 {
		writeArgs(writer, cmd.Args)
	}

	// writes flags sorting them first by required/not required and then alphabetically
	if len(cmd.Flags) > 0 {
		sort.Slice(cmd.Flags, func(i, j int) bool {
			result := cmd.Flags[i].IsRequired()
			if cmd.Flags[i].IsRequired() == cmd.Flags[j].IsRequired() {
				result = cmd.Flags[i].GetName() < cmd.Flags[j].GetName()
			}
			return result
		})
		writeFlags(writer, cmd.Flags)
	}

	// includes any extra markdown command examples
	examplesMdFile, err := getMarkdownFileContent(fullCmdName, examplesFile)
	if err != nil {
		return err
	}

	if examplesMdFile != "" {
		fmt.Fprintf(writer, "%s \n\n", markdown.WrapBold("Examples:"))
		fmt.Fprintf(writer, "%s \n\n", examplesMdFile)
	}

	return nil
}

// writeArgs writes args section to a markdown table
func writeArgs(writer io.Writer, cmdArgs cli.Args) {
	fmt.Fprintf(writer, "%s: \n\n", markdown.WrapBold("Args"))

	markdown.WriteTableHeader(writer, []string{"Name", "Description"})

	for _, tempArg := range cmdArgs {
		writeArgTableRow(writer, tempArg)
	}

	fmt.Fprintln(writer, "")
}

// writeArgTableRow writes arg table row
func writeArgTableRow(w io.Writer, arg cli.Arg) {
	name := markdown.WrapInlineCode(arg.Name)

	if arg.Required {
		name += fmt.Sprintf(" %s", markdown.WrapBold("(required)"))
	}

	row := []string{name, arg.Description}

	markdown.WriteTableRow(w, row)
}

// writeFlags writes flags section to a markdown table
func writeFlags(writer io.Writer, cmdFlags cli.Flags) {
	fmt.Fprintf(writer, "%s \n\n", markdown.WrapBold("Flags:"))

	markdown.WriteTableHeader(writer, []string{"Name", "Description"})

	for _, tempFlag := range cmdFlags {
		writeFlagTableRow(writer, tempFlag)
	}

	fmt.Fprintln(writer, "")
}

// writeFlagTableRow writes flag table row
func writeFlagTableRow(w io.Writer, flag cli.Flag) {
	name := fmt.Sprintf("--%s", flag.GetName())

	// if the flag isn't a bool flag then print out a placeholder value with the name of the flag
	if _, ok := flag.(*cli.BoolFlag); !ok {
		name += fmt.Sprintf(" %s", strings.ToUpper(flag.GetName()))
	}
	name = markdown.WrapInlineCode(name)

	short := flag.GetShort()
	if short != "" {
		name += fmt.Sprintf(", %s", markdown.WrapInlineCode(fmt.Sprintf("-%s", short)))
	}

	if flag.IsRequired() {
		name += fmt.Sprintf(" %s", markdown.WrapBold("(required)"))
	}

	row := []string{name, flag.GetUsage()}

	markdown.WriteTableRow(w, row)
}

// getMarkdownFileContent reads external Markdown file content
func getMarkdownFileContent(cmdName string, fileName string) (string, error) {
	// root path
	markdownPath := InputDocsDir

	// appends the remainder of the file path
	words := strings.Split(cmdName, ":")
	for _, subCmd := range words {
		markdownPath += "/" + subCmd
	}

	// appends file name
	markdownPath += "/" + fileName

	buffer, err := ioutil.ReadFile(markdownPath)

	if err != nil {
		if !os.IsNotExist(err) {
			return "", err
		}
		// File doesn't exist, so nothing to return
		return "", nil
	}

	log.Printf("Markdown file found: %s", markdownPath)
	return string(buffer), nil
}
