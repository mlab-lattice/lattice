package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"

	"github.com/mlab-lattice/lattice/cmd/latticectl/app"
	"github.com/mlab-lattice/lattice/pkg/util/cli/docgen"
)

func main() {
	// reading flags from command line
	inputDocsDirPtr := flag.String("input-docs", "./docs/cli", "Extra markdown docs input directory")
	outputDocsDirPtr := flag.String("output-docs", "./doc.md", "Markdown docs output directory")

	flag.Parse()

	inputDocsDir := *inputDocsDirPtr
	docgen.InputDocsDir = *inputDocsDirPtr
	outputDocsDir := *outputDocsDirPtr

	log.Printf("Input docs dir: '%s' \n", inputDocsDir)
	log.Printf("Output docs dir: '%s' \n", outputDocsDir)

	latticectl := app.Latticectl
	cmd, er := latticectl.Init()
	if er != nil {
		log.Fatalf("FATAL: Error while initialising laasctl")
	}

	reader, err := docgen.GenerateMarkdown(cmd)
	if err != nil {
		log.Fatalf("FATAL: Error while generating markdown: %s", err)
	}

	// opens docs output file
	fo, err := os.Create(outputDocsDir)
	if err != nil {
		log.Fatalf("FATAL: Error while creating doc markdown file: %s", err)
	}

	// closes docs output file on exit
	defer func() error {
		if err := fo.Close(); err != nil {
			return err
		}
		return nil
	}()

	// writes markdown to the file
	markdownBytes, err := ioutil.ReadAll(reader)
	fo.Write(markdownBytes)
	fo.Close()
}
