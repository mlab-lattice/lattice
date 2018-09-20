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
	"github.com/mlab-lattice/lattice/pkg/util/cli/flags"
	"github.com/mlab-lattice/lattice/pkg/util/markdown"
)

var InputDocsDir string

// extra markdown file name
const (
	descriptionFile = "description.md"
	examplesFile    = "examples.md"
)

type flagInfo struct {
	flag cli.Flag
	name string
}

func GenerateMarkdown(cmd *cli.RootCommand) (io.Reader, error) {
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
func writeDoc(bc *cli.RootCommand, writer io.Writer) error {
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
	for name, childCmd := range bc.Subcommands {
		var cmds []string
		recurse(name, childCmd, cmds, writer, commandMapping)
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
func recurse(name string, cmd *cli.Command, ancestorCommands []string, writer io.Writer, commandMapping map[string]*cli.Command) {
	if name == "" {
		return
	}

	// joins consecutive ancestor command names
	ancestorCmdsStr := strings.Join(ancestorCommands, " ")
	ancestorCmdsStr += name

	commandMapping[ancestorCmdsStr] = cmd

	// adds this command to the list of ancestor commands to be passed to its children when recursing
	ancestorCommands = append(ancestorCommands, name)

	for name, subcmd := range cmd.Subcommands {
		recurse(name, subcmd, ancestorCommands, writer, commandMapping)
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

		var info []flagInfo
		for name, flag := range cmd.Flags {
			info = append(info, flagInfo{flag, name})
		}

		sort.Slice(info, func(i, j int) bool {
			if info[i].flag.IsRequired() == info[j].flag.IsRequired() {
				return info[i].name < info[j].name
			}

			return info[i].flag.IsRequired()
		})
		writeFlags(writer, info)
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
func writeFlags(writer io.Writer, info []flagInfo) {
	fmt.Fprintf(writer, "%s \n\n", markdown.WrapBold("Flags:"))

	markdown.WriteTableHeader(writer, []string{"Name", "Description"})

	for _, flagInfo := range info {
		writeFlagTableRow(writer, flagInfo.name, flagInfo.flag)
	}

	fmt.Fprintln(writer, "")
}

// writeFlagTableRow writes flag table row
func writeFlagTableRow(w io.Writer, flagName string, flag cli.Flag) {
	name := fmt.Sprintf("--%s", flagName)

	// if the flag isn't a bool flag then print out a placeholder value with the name of the flag
	if _, ok := flag.(*flags.Bool); !ok {
		name += fmt.Sprintf(" %s", strings.ToUpper(flagName))
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
