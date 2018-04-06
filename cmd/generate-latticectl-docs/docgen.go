package main

import (
	"flag"
	"github.com/mlab-lattice/lattice/pkg/util/cli/docgen"
	"log"
	"os"
	"github.com/mlab-lattice/lattice/cmd/latticectl/definition"
)


func main() {
	// reading flags from command line
	projectDir := os.Getenv("GOPATH") + "/src/github.com/mlab-lattice/lattice"
	inputDocsDirPtr := flag.String("input-docs", projectDir, "Extra markdown docs input directory")
	outputDocsDirPtr := flag.String("output-docs", projectDir, "Markdown docs output directory")

	flag.Parse()

	inputDocsDir := *inputDocsDirPtr
	docgen.InputDocsDir = *inputDocsDirPtr
	outputDocsDir := *outputDocsDirPtr

	log.Printf("Input docs dir: '%s' \n", inputDocsDir)
	log.Printf("Output docs dir: '%s' \n", outputDocsDir)

	latticectl := definition.GenerateLatticeCtl()
	cmd, er := latticectl.Init()

	if er != nil {
		log.Fatalf("FATAL: Error while initialising laasctl")
	}

	docgen.GenerateCtlDoc(cmd, *outputDocsDirPtr)
}