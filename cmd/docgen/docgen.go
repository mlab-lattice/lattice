package main

import (
	"io/ioutil"
	"log"
	"plugin"

	"github.com/mlab-lattice/lattice/pkg/util/cli"
	"github.com/mlab-lattice/lattice/pkg/util/cli/docgen"
	"github.com/mlab-lattice/lattice/pkg/util/cli/flags"
)

func main() {
	Command().Execute()
}

func Command() *cli.RootCommand {
	var (
		inputDir   string
		outputDir  string
		pluginPath string
	)

	return &cli.RootCommand{
		Name: "docgen",
		Command: &cli.Command{
			Flags: cli.Flags{
				"input-docs": &flags.String{
					Default: "./docs/cli",
					Usage:   "extra markdown docs input directory",
					Target:  &inputDir,
				},
				"output-docs": &flags.String{
					Default: "./doc.md",
					Usage:   "markdown docs output file path",
					Target:  &outputDir,
				},
				"plugin": &flags.String{
					Usage:    "path to plugin file containing the command to generate documentation for",
					Required: true,
					Target:   &pluginPath,
				},
			},
			Run: func(args []string, flags cli.Flags) error {
				log.Printf("Input docs dir: '%s' \n", inputDir)
				log.Printf("Output docs file path: '%s' \n", outputDir)

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

				reader, err := docgen.GenerateMarkdown(command)
				if err != nil {
					log.Fatalf("FATAL: Error while generating markdown: %s", err)
				}

				markdownBytes, err := ioutil.ReadAll(reader)
				if err != nil {
					log.Fatalf("FATAL: Error while reading from markdown buffer: %s", err)
				}

				writeError := ioutil.WriteFile(outputDir, markdownBytes, 0755)
				if writeError != nil {
					log.Fatalf("FATAL: Error while writing markdown buffer to file: %s", err)
				}
				return nil
			},
		},
	}
}
