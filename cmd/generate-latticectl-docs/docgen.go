package main

import (
	"flag"
	"io/ioutil"
	"log"

	"github.com/mlab-lattice/lattice/cmd/latticectl/app"
	"github.com/mlab-lattice/lattice/pkg/util/cli/docgen"
)

func main() {
	// reading flags from command line
	inputDocsDirPtr := flag.String("input-docs", "./docs/cli", "Extra markdown docs input directory")
	outputDocsFilePathPtr := flag.String("output-docs", "./doc.md", "Markdown docs output file path")

	flag.Parse()

	inputDocsDir := *inputDocsDirPtr
	docgen.InputDocsDir = *inputDocsDirPtr
	outputDocsFilePath := *outputDocsFilePathPtr

	log.Printf("Input docs dir: '%s' \n", inputDocsDir)
	log.Printf("Output docs file path: '%s' \n", outputDocsFilePath)

	latticectl := app.Latticectl
	cmd, er := latticectl.Init()
	if er != nil {
		log.Fatalf("FATAL: Error while initialising laasctl")
	}

	reader, err := docgen.GenerateMarkdown(cmd)
	if err != nil {
		log.Fatalf("FATAL: Error while generating markdown: %s", err)
	}

	markdownBytes, err := ioutil.ReadAll(reader)
	if err != nil {
		log.Fatalf("FATAL: Error while reading from markdown buffer: %s", err)
	}

	writeError := ioutil.WriteFile(outputDocsFilePath, markdownBytes, 0755)
	if writeError != nil {
		log.Fatalf("FATAL: Error while writing markdown buffer to file: %s", err)
	}
}
