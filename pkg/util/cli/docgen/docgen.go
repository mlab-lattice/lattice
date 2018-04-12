package docgen

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"reflect"
	"sort"
	"strings"

	"github.com/mlab-lattice/lattice/pkg/util/cli"
	"github.com/mlab-lattice/lattice/pkg/util/markdown"
)

var InputDocsDir string

func GenerateMarkdown(cmd *cli.Command) (io.Reader, error) {
	var buffer bytes.Buffer
	writer := bufio.NewWriter(&buffer)

	writeErr := writeDoc(cmd, writer)
	if writeErr != nil {
		return nil, writeErr
	}

	writer.Flush()

	return bytes.NewReader(buffer.Bytes()), nil
}

// writeDoc writes command tree to file
func writeDoc(bc *cli.Command, writer io.Writer) error {
	// header
	fmt.Fprint(writer, markdown.GetH1(bc.Name))
	fmt.Fprint(writer, markdown.GetH2("Introduction"))
	fmt.Fprint(writer, bc.Short+"\n")
	fmt.Fprint(writer, markdown.GetH2("Commands"))

	// mapping between command name (e.g. `laasctl systems builds`) and the command.
	// the output to file can therefore be sorted by command name (ascending),
	// while still using depth-first tree traversal
	commandMapping := make(map[string]*cli.Command)

	// reads the tree and appends commands into the commandMapping
	for _, childCmd := range bc.Subcommands {
		var cmds []cli.Command
		recurse(childCmd, cmds, writer, commandMapping)
	}

	// uses sorted map keys to iterate over the map in alphabetical
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
		ancestorCmdsStr += tempCmd.Name + " "
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
	// header command name (add extra #s here if want hierarchy)
	fmt.Fprint(writer, markdown.GetH3(fullCmdName))

	// description
	if cmd.Short != "" {
		fmt.Fprint(writer, "*"+cmd.Short+"*  \n\n")
	}

	// includes any extra markdown command description
	markdownContent, err := getMarkdownFileContent(fullCmdName)
	if err != nil {
		return err
	}

	if markdownContent != "" {
		fmt.Fprint(writer, markdownContent)
	}

	// validates that required args are defined first
	if len(cmd.Args) > 0 {
		writeArgs(writer, cmd.Args)
	}

	// write flags sorting them first by required/not required and then alphabetically
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

	return nil
}

// writeArgs writes args section to a markdown table
func writeArgs(writer io.Writer, cmdArgs cli.Args) {
	fmt.Fprint(writer, "**Args**:  \n\n")

	// table header
	fmt.Fprint(writer, "| Name | Description | \n")
	fmt.Fprint(writer, "| --- | --- | \n")

	// table contents
	for _, tempArg := range cmdArgs {
		argRow := getArgTableRow(tempArg)
		fmt.Fprint(writer, argRow)
	}

	// new line separator
	fmt.Fprint(writer, "\n\n")
}

// writeFlags writes flags section to a markdown table
func writeFlags(writer io.Writer, cmdFlags cli.Flags) {
	fmt.Fprint(writer, "**Flags**:  \n\n")

	// table header
	fmt.Fprint(writer, "| Name | Description | \n")
	fmt.Fprint(writer, "| --- | --- | \n")

	// table contents
	for _, tempFlag := range cmdFlags {
		flagRow := getFlagTableRow(tempFlag)
		fmt.Fprint(writer, flagRow)
	}

	// new line separator
	fmt.Fprint(writer, "\n\n")
}

// getFlagTableRow generates flag table row
func getFlagTableRow(flagPtr cli.Flag) string {
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

	return flagRow
}

// getArgTableRow generates arg table row
func getArgTableRow(argPtr cli.Arg) string {
	currentArg := argPtr

	argRow := "| `" + currentArg.Name + "` "

	if currentArg.Required {
		argRow += "**(required)** "
	}

	argRow += "| "

	argRow += currentArg.Description + " | \n"

	return argRow
}

// getMarkdownFileContent reads external Markdown file content
func getMarkdownFileContent(cmdName string) (string, error) {

	// root path
	markdownPath := InputDocsDir

	// appends the remainder of the file path
	words := strings.Fields(cmdName)
	for _, subCmd := range words {
		markdownPath += "/" + subCmd
	}

	// appends file name
	markdownPath += "/description.md"

	buffer, err := ioutil.ReadFile(markdownPath)

	if err == nil {
		log.Printf("Markdown file found: %s", markdownPath)
		return string(buffer) + "\n", nil
	} else if strings.Contains(err.Error(), "no such file or directory") {
		return "", nil
	}

	log.Printf("Error: Markdown file '%s' cannot be read: %s", markdownPath, err)
	return "", err
}
