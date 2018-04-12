package docgen

import (
	"bufio"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"sort"
	"strings"
)

var InputDocsDir string


func GenerateMarkdown(cmd *cli.Command, outputDir string) error{
	// opens docs output file
	fo, err := os.Create(outputDir + "/doc.md")
	if err != nil {
		return err
	}

	// closes docs output file on exit
	defer func() error {
		if err := fo.Close(); err != nil {
			return err
		}
		return nil
	}()

	writer := bufio.NewWriter(fo)

	writeErr := writeDoc(cmd, writer)
	if writeErr != nil {
		return writeErr
	}

	writer.Flush()

	return nil
}


// writeDoc writes command tree to file
func writeDoc(bcPtr *cli.Command, writer *bufio.Writer) error {
	bc := *bcPtr

	// header
	writer.WriteString("# " + bc.Name + "\n")
	writer.WriteString("## Introduction \n")
	writer.WriteString(bc.Short + "\n")
	writer.WriteString("## Commands \n")

	// mapping between command name (e.g. `laasctl systems builds`) and the command.
	// the output to file can therefore be sorted by command name (ascending),
	// while still using depth-first tree traversal
	commandMapping := make(map[string]*cli.Command)

	// reads the tree and appends commands into the commandMapping
	for _, childCmd := range bc.Subcommands {
		var cmds []cli.Command
		recurse(childCmd, &cmds, writer, &commandMapping)
	}

	// uses sorted map keys to iterate over the map in alphabetical
	keys := make([]string, 0)
	for k := range commandMapping {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, key := range keys {
		err := printCommand(&key, commandMapping[key], writer)
		if err != nil {
			return err
		}
	}

	return nil
}


// recurse recursively works through the command tree
// ancestorCommands are commands that precede the current cmd in the hierarchy
func recurse(cmdPtr *cli.Command, ancestorCommandsPtr *[]cli.Command, writer *bufio.Writer, commandMappingPtr *map[string]*cli.Command) {

	ancestorCommands := *ancestorCommandsPtr
	cmd := *cmdPtr
	commandMapping := *commandMappingPtr

	if cmd.Name == "" {
		return
	}

	// joins consecutive ancestor command names
	var ancestorCmdsStr string
	for _, tempCmd := range ancestorCommands {
		ancestorCmdsStr += tempCmd.Name + " "
	}
	ancestorCmdsStr += cmd.Name

	commandMapping[ancestorCmdsStr] = cmdPtr

	// adds this command to the list of ancestor commands to be passed to its children when recursing
	ancestorCommands = append(ancestorCommands, cmd)

	for _, tempCmd := range cmd.Subcommands {
		recurse(tempCmd, &ancestorCommands, writer, commandMappingPtr)
	}
}


// printCommand prints the command docs
// fullCmdName includes all command ancestor except the root command (e.g. 'latticectl')
func printCommand(fullCmdName *string, cmdPtr *cli.Command, writer *bufio.Writer) error {
	cmd := *cmdPtr

	// header command name (add extra #s here if want hierarchy)
	writer.WriteString("### " + *fullCmdName + "  \n")

	// description
	if cmd.Short != "" {
		writer.WriteString("*" + cmd.Short + "*  \n\n")
	}

	// includes any extra markdown command description
	markdownContent, err := getMarkdownFileContent(fullCmdName)
	if err != nil {
		return err
	}

	if markdownContent != "" {
		writer.WriteString(markdownContent)
	}

	// validates that required args are defined first
	if len(cmd.Args) > 0 {
		writeArgs(writer, &cmd.Args)
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
		writeFlags(writer, &cmd.Flags)
	}

	return nil
}


// writeArgs writes args section to a markdown table
func writeArgs(writer *bufio.Writer, cmdArgs *cli.Args) {
	writer.WriteString("**Args**:  \n\n")

	// table header
	writer.WriteString("| Name | Description | \n")
	writer.WriteString("| --- | --- | \n")

	// table contents
	for _, tempArg := range *cmdArgs {
		argRow := getArgTableRow(&tempArg)
		writer.WriteString(argRow)
	}

	// new line separator
	writer.WriteString("\n\n")
}


// writeFlags writes flags section to a markdown table
func writeFlags(writer *bufio.Writer, cmdFlags *cli.Flags) {
	writer.WriteString("**Flags**:  \n\n")

	// table header
	writer.WriteString("| Name | Description | \n")
	writer.WriteString("| --- | --- | \n")

	// table contents
	for _, tempFlag := range *cmdFlags {
		flagRow := getFlagTableRow(&tempFlag)
		writer.WriteString(flagRow)
	}

	// new line separator
	writer.WriteString("\n\n")
}


// getFlagTableRow generates flag table row
func getFlagTableRow(flagPtr *cli.Flag) string {
	currentFlag := *flagPtr
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
func getArgTableRow(argPtr *cli.Arg) string {
	currentArg := *argPtr

	argRow := "| `" + currentArg.Name + "` "

	if currentArg.Required {
		argRow += "**(required)** "
	}

	argRow += "| "

	argRow += currentArg.Description + " | \n"

	return argRow
}


// getMarkdownFileContent reads external Markdown file content
func getMarkdownFileContent(cmdName *string) (string, error) {

	// root path
	markdownPath := InputDocsDir

	// appends the remainder of the file path
	words := strings.Fields(*cmdName)
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
