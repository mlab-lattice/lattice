package main

import (
	"flag"
	"github.com/mlab-lattice/lattice/pkg/util/cli/docgen"
	"log"
	"github.com/mlab-lattice/lattice/cmd/latticectl/app"
)


func main() {
	// reading flags from command line
	inputDocsDirPtr := flag.String("input-docs", "./docs/cli", "Extra markdown docs input directory")
	outputDocsDirPtr := flag.String("output-docs", ".", "Markdown docs output directory")

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

	docgen.GenerateMarkdown(cmd, *outputDocsDirPtr)
}