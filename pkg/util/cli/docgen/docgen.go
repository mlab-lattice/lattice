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

/**
Generates laasctl docs
*/
func GenerateCtlDoc(cmd *cli.Command, outputDir string) {
	// open file
	fo, err := os.Create(outputDir + "/doc.md")
	if err != nil {
		panic(err)
	}
	// close fo on exit and check for its returned error
	defer func() {
		if err := fo.Close(); err != nil {
			panic(err)
		}
	}()

	writer := bufio.NewWriter(fo)

	printIntro(cmd, writer)
	writer.Flush()
}

/**
Prints headings and intro
*/
func printIntro(bcPtr *cli.Command, writer *bufio.Writer) {
	bc := *bcPtr

	// header
	writer.WriteString("# " + bc.Name + "\n")
	writer.WriteString("## Introduction \n")
	writer.WriteString(bc.Short + "\n")
	writer.WriteString("## Commands \n")

	// mapping between command name (e.g. `lattices status` and the command)
	commandMapping := make(map[string]*cli.Command)

	// reads the tree and appends commands into the commandMapping
	for _, tempCmd := range bc.Subcommands {
		cmds := []cli.Command{bc}
		recurse(tempCmd, &cmds, writer, &commandMapping)
	}

	// sorts map keys ASC
	keys := make([]string, 0)
	for k := range commandMapping {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	// uses sort map keys to iterate over the map
	for _, key := range keys {
		printCommand(&key, commandMapping[key], writer)
	}
}

/**
Recursively works through the command tree
*/
func recurse(cmdPtr *cli.Command, allCommandsPtr *[]cli.Command, writer *bufio.Writer, commandMappingPtr *map[string]*cli.Command) {

	allCommands := *allCommandsPtr
	cmd := *cmdPtr
	commandMapping := *commandMappingPtr

	if cmd.Name == "" {
		return
	}

	mostCommands := allCommands[1:] // all commands but the first one (`lattice`)

	// get full command except (lattice) as a string
	var prevCommands string

	for _, tempCmd := range mostCommands {
		prevCommands += tempCmd.Name + " "
	}

	prevCommands += cmd.Name

	// add new element to the map
	commandMapping[prevCommands] = cmdPtr

	allCommands = append(allCommands, cmd)

	for _, tempCmd := range cmd.Subcommands {
		recurse(tempCmd, &allCommands, writer, commandMappingPtr)
	}
}

/**
Prints the command docs
*/
func printCommand(prevCmdName *string, cmdPtr *cli.Command, writer *bufio.Writer) {
	cmd := *cmdPtr

	// header command name (add extra #s here if want hierarchy)
	writer.WriteString("### ")
	writer.WriteString(*prevCmdName + "  \n")

	// description
	if cmd.Short != "" {
		writer.WriteString("*" + cmd.Short + "*  \n\n")
	}

	// markdown import
	markdownContent := getMarkdownFileContent(prevCmdName)
	if markdownContent != "" {
		writer.WriteString(markdownContent)
	}

	// write args
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

	// write flags
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

/**
Writes args to a markdown table
*/
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

/**
Writes flags to a markdown table
*/
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

/**
Generates a string for the flag table row
*/
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

/**
Generates a string for the arg table row
*/
func getArgTableRow(argPtr *cli.Arg) string {
	currentArg := *argPtr

	// name
	argRow := "| `" + currentArg.Name + "` "

	// is required
	if currentArg.Required {
		argRow += "**(required)** "
	}

	argRow += "| "

	// mock description until the description field in the struct is available
	argRow += currentArg.Description + " | \n"

	return argRow
}

/**
Returns external Markdown file content
*/
func getMarkdownFileContent(previousCommmands *string) string {

	// root path
	markDownPath := InputDocsDir + "/docs/cli"

	// splits string by whitespace
	words := strings.Fields(*previousCommmands)

	for _, subCmd := range words {
		markDownPath += "/" + subCmd
	}

	// end of path
	markDownPath += "/description.md"

	buffer, err := ioutil.ReadFile(markDownPath)

	if err == nil {
		log.Printf("Markdown file found: %s", markDownPath)
		return string(buffer) + "\n"
	}

	return ""
}
