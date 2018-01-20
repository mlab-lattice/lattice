package definition

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/mlab-lattice/system/pkg/cli"
	definitionresolver "github.com/mlab-lattice/system/pkg/definition/resolver"
	"github.com/mlab-lattice/system/pkg/util/git"

	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:  "definition",
	Args: cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
		os.Exit(1)
	},
}

var versionsCmd = &cobra.Command{
	Use:   "versions [url]",
	Short: "list the versions of a definition",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		definitionURL := args[0]

		resolver := getResolver()

		versions, err := resolver.ListDefinitionVersions(definitionURL, &git.Options{})
		if err != nil {
			fmt.Printf("error retrieving system definition versions: %v\n", err)
		}

		cli.ListSystemDefinitionVersions(versions)
	},
}

var resolveCmd = &cobra.Command{
	Use:   "resolve [url] [version]",
	Short: "resolve a definition",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		definitionURL := args[0]
		version := args[1]

		resolver := getResolver()

		uri := fmt.Sprintf("%v#%v/system.json", definitionURL, version)
		definition, err := resolver.ResolveDefinition(uri, &git.Options{})
		if err != nil {
			fmt.Printf("error retrieving system definition %v: %v\n", uri, err)
			os.Exit(1)
		}

		definitionJSON, err := json.MarshalIndent(definition, "", "  ")
		if err != nil {
			fmt.Printf("error unmarshalling definition json: %v\n", err)
			os.Exit(1)
		}

		fmt.Println(string(definitionJSON))
	},
}

func getResolver() *definitionresolver.SystemResolver {
	resolver, err := definitionresolver.NewSystemResolver("/tmp/latticectl/definition")
	if err != nil {
		fmt.Printf("error creating system definition resolver: %v\n", err)
		os.Exit(1)
	}

	return resolver
}

func init() {
	Cmd.AddCommand(versionsCmd)
	Cmd.AddCommand(resolveCmd)
}
