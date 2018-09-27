package main

import (
	"io"
	"log"
	"os"
	"plugin"

	//"fmt"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
	"github.com/mlab-lattice/lattice/pkg/util/cli/docgen"
	"github.com/mlab-lattice/lattice/pkg/util/cli/flags"
)

func main() {
	Command().Execute()
}

const (
	extraMarkdownFlag = "extra-markdown"
)

func Command() *cli.RootCommand {
	var (
		extraMarkdown string
		pluginPath    string
	)

	return &cli.RootCommand{
		Name: "docgen",
		Command: &cli.Command{
			Args: cli.Args{
				AllowAdditional: true,
			},
			Flags: cli.Flags{
				extraMarkdownFlag: &flags.String{
					Usage:  "path to extra markdown to be used when generating documentation",
					Target: &extraMarkdown,
				},
				"plugin": &flags.String{
					Usage:    "path to plugin file containing the command to generate documentation for",
					Required: true,
					Target:   &pluginPath,
				},
			},
			Run: func(args []string, flags cli.Flags) error {
				p, err := plugin.Open(pluginPath)
				if err != nil {
					return err
				}

				f, err := p.Lookup("Command")
				if err != nil {
					return err
				}

				command := f.(func() *cli.RootCommand)()
				err = command.Init()
				if err != nil {
					log.Fatalf("FATAL: Error while initialising latticectl")
				}

				generator := docgen.NewGenerator(command, extraMarkdown)
				reader, err := generator.Markdown()
				if err != nil {
					log.Fatalf("FATAL: Error while generating markdown: %s", err)
				}

				_, err = io.Copy(os.Stdout, reader)
				return err
			},
		},
	}
}
