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


func GenerateCliDocs(cmd *cli.Command, outputDir string) {
	// opens docs output file
	fo, err := os.Create(outputDir + "/doc.md")
	if err != nil {
		panic(err)
	}
	// closes docs output file on exit
	defer func() {
		if err := fo.Close(); err != nil {
			panic(err)
		}
	}()

	writer := bufio.NewWriter(fo)

	writeDoc(cmd, writer)
	writer.Flush()
}


// writeDoc writes command tree to file
func writeDoc(bcPtr *cli.Command, writer *bufio.Writer) {
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
	for _, tempCmd := range bc.Subcommands {
		cmds := []cli.Command{bc}
		recurse(tempCmd, &cmds, writer, &commandMapping)
	}

	// uses sorted map keys to iterate over the map in alphabetical
	keys := make([]string, 0)
	for k := range commandMapping {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, key := range keys {
		printCommand(&key, commandMapping[key], writer)
	}
}


// recurse recursively works through the command tree
func recurse(cmdPtr *cli.Command, allCommandsPtr *[]cli.Command, writer *bufio.Writer, commandMappingPtr *map[string]*cli.Command) {

	allCommands := *allCommandsPtr
	cmd := *cmdPtr
	commandMapping := *commandMappingPtr

	if cmd.Name == "" {
		return
	}

	// all commands but the first one (e.g. `latticectl`)
	mostCommands := allCommands[1:]

	// retrieves the above as a string
	var prevCommands string
	for _, tempCmd := range mostCommands {
		prevCommands += tempCmd.Name + " "
	}
	prevCommands += cmd.Name

	// adds current command to the command mapping
	commandMapping[prevCommands] = cmdPtr

	// forms full command slice (including current command)
	allCommands = append(allCommands, cmd)

	for _, tempCmd := range cmd.Subcommands {
		recurse(tempCmd, &allCommands, writer, commandMappingPtr)
	}
}


// printCommand prints the command docs
func printCommand(cmdName *string, cmdPtr *cli.Command, writer *bufio.Writer) {
	cmd := *cmdPtr

	// header command name (add extra #s here if want hierarchy)
	writer.WriteString("### " + *cmdName + "  \n")

	// description
	if cmd.Short != "" {
		writer.WriteString("*" + cmd.Short + "*  \n\n")
	}

	// includes any extra markdown command description
	markdownContent := getMarkdownFileContent(cmdName)
	if markdownContent != "" {
		writer.WriteString(markdownContent)
	}

	// write args sorting them first by required/not required and then alphabetically
	if len(cmd.Args) > 0 {
		sort.Slice(cmd.Args, func(i, j int) bool {
			result := cmd.Args[i].Required
			if cmd.Args[i].Required == cmd.Args[j].Required {
				result = cmd.Args[i].Name < cmd.Args[j].Name
			}
			return result
		})
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
func getMarkdownFileContent(cmdName *string) string {

	// root path
	markdownPath := InputDocsDir + "/docs/cli"

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
		return string(buffer) + "\n"
	}

	return ""
}
