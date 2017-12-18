package cli

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/mlab-lattice/system/pkg/cli/resources"
	"github.com/olekukonko/tablewriter"
)

func showResource(resource resources.EndpointResource) {
	renderMap := resource.GetRenderMap()
	table := tablewriter.NewWriter(os.Stdout)
	table.SetRowLine(true)
	for k, v := range renderMap {
		table.Append([]string{strings.ToUpper(k), v})
	}
	table.Render()
}

func listResources(resources []resources.EndpointResource) {
	if len(resources) > 0 {
		keys := []string{}
		for k := range resources[0].GetRenderMap() {
			keys = append(keys, k)
		}
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader(keys)
		for _, r := range resources {
			m := r.GetRenderMap()
			line := make([]string, 0, len(keys))
			for _, k := range keys {
				line = append(line, m[k])
			}
			table.Append(line)
		}
		table.Render()
	}
}

func DisplayAsJSON(resource interface{}) {
	buf, err := json.MarshalIndent(resource, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(buf))
}
