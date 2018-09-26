package docgen

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sort"
	"strings"

	"github.com/mlab-lattice/lattice/pkg/util/cli"
	"github.com/mlab-lattice/lattice/pkg/util/cli/flags"
	"github.com/mlab-lattice/lattice/pkg/util/markdown"
	"path/filepath"
)

// extra markdown file name
const (
	descriptionFile = "description.md"
	examplesFile    = "examples.md"
)

func NewGenerator(cmd *cli.RootCommand, externalMarkdownRoot string) *Generator {
	return &Generator{cmd, externalMarkdownRoot}
}

type Generator struct {
	cmd                  *cli.RootCommand
	externalMarkdownRoot string
}

type flagInfo struct {
	flag cli.Flag
	name string
}

func (g *Generator) Markdown() (io.Reader, error) {
	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)

	err := g.write(w)
	if err != nil {
		return nil, err
	}

	w.Flush()
	return bytes.NewReader(buf.Bytes()), nil
}

// write writes command tree to file
func (g *Generator) write(w io.Writer) error {
	// header
	fmt.Fprintf(w, "%s \n", markdown.WrapH1(g.cmd.Name))
	fmt.Fprintf(w, "%s \n", markdown.WrapH2("Introduction"))

	// extra description in the intro section
	err := g.writeExternalDescription(w, "")
	if err != nil {
		return err
	}

	fmt.Fprintf(w, "%s \n", markdown.WrapH2("Commands"))

	// mapping between command name (e.g. `latticectl systems status`) and the command.
	// the output to file can therefore be sorted by command name (ascending),
	// while still using depth-first tree traversal
	commands := make(map[string]*cli.Command)

	// reads the tree and adds commands into the commands
	for name, cmd := range g.cmd.Subcommands {
		var path []string
		g.commands(name, cmd, path, commands)
	}

	// uses sorted command names to iterate over the commands in alphabetical order
	var names []string
	for k := range commands {
		names = append(names, k)
	}
	sort.Strings(names)

	for _, name := range names {
		err := g.writeCommand(w, name, commands[name])
		if err != nil {
			return err
		}
	}

	return nil
}

// commands recursively works through the command tree
// path is a list of the ancestor command names to get to the current command
func (g *Generator) commands(name string, cmd *cli.Command, path []string, commandMapping map[string]*cli.Command) {
	fullPath := fmt.Sprintf("%v %v", strings.Join(path, " "), name)
	commandMapping[fullPath] = cmd

	// adds this command to the list of ancestor commands to be passed to its children when recursing
	path = append(path, name)
	for name, subcmd := range cmd.Subcommands {
		g.commands(name, subcmd, path, commandMapping)
	}
}

// writeCommand prints the command docs
// fullName includes all command ancestor except the root command (e.g. 'latticectl')
func (g *Generator) writeCommand(w io.Writer, name string, cmd *cli.Command) error {
	fmt.Fprintf(w, "%s  \n", markdown.WrapH2(name))

	if cmd.Short != "" {
		fmt.Fprintf(w, "%s  \n\n", markdown.WrapItalic(cmd.Short))
	}

	// includes any extra markdown command description
	err := g.writeExternalDescription(w, name)
	if err != nil {
		return err
	}

	g.writeArgs(w, cmd.Args)
	g.writeFlags(w, cmd)

	// includes any extra markdown command examples
	err = g.writeExternalExamplesMarkdown(w, name)
	if err != nil {
		return err
	}

	return nil
}

// writeArgs writes args section to a markdown table
func (g *Generator) writeArgs(w io.Writer, args cli.Args) {
	// TODO: handle args.AllowAdditional
	if len(args.Args) == 0 {
		return
	}

	fmt.Fprintf(w, "%s: \n\n", markdown.WrapBold("Args"))

	markdown.WriteTableHeader(w, []string{"Name", "Description"})

	for _, tempArg := range args.Args {
		writeArgTableRow(w, tempArg)
	}

	fmt.Fprintln(w, "")
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
func (g *Generator) writeFlags(w io.Writer, cmd *cli.Command) {
	if len(cmd.Flags) == 0 {
		return
	}

	// sort first by required/not required and then alphabetically
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

	fmt.Fprintf(w, "%s \n\n", markdown.WrapBold("Flags:"))

	markdown.WriteTableHeader(w, []string{"Name", "Description"})

	for _, flagInfo := range info {
		writeFlagTableRow(w, flagInfo.name, flagInfo.flag)
	}

	fmt.Fprintln(w, "")
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

func (g *Generator) writeExternalDescription(w io.Writer, command string) error {
	desc, err := g.externalDescriptionMarkdown(command)
	if err != nil {
		return err
	}

	if desc != "" {
		fmt.Fprintf(w, "%s \n\n", desc)
	}

	return nil
}

func (g *Generator) writeExternalExamplesMarkdown(w io.Writer, command string) error {
	examples, err := g.externalExamplesMarkdown(command)
	if err != nil {
		return err
	}

	if examples != "" {
		fmt.Fprintf(w, "%s \n\n", markdown.WrapBold("Examples:"))
		fmt.Fprintf(w, "%s \n\n", examples)
	}

	return nil
}

func (g *Generator) externalDescriptionMarkdown(command string) (string, error) {
	if g.externalMarkdownRoot == "" {
		return "", nil
	}

	file := g.externalMarkdownPath(command, descriptionFile)
	return g.externalMarkdown(file)
}

func (g *Generator) externalExamplesMarkdown(command string) (string, error) {
	if g.externalMarkdownRoot == "" {
		return "", nil
	}

	file := g.externalMarkdownPath(command, examplesFile)
	return g.externalMarkdown(file)
}

func (g *Generator) externalMarkdownPath(command, file string) string {
	parts := strings.Split(command, " ")
	path := filepath.Join(parts...)
	return filepath.Join(g.externalMarkdownRoot, path, file)
}

// externalMarkdown reads external Markdown file content
func (g *Generator) externalMarkdown(file string) (string, error) {
	buf, err := ioutil.ReadFile(file)
	if err != nil {
		// it's okay for a file to not exist, so swallow the error
		// using this type check instead of os.IsNotExist to swallow
		// both the file not existing or the directory not existing
		if _, ok := err.(*os.PathError); ok {
			return "", nil
		}

		return "", nil
	}

	return string(buf), nil
}
