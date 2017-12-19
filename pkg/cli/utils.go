package cli

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/olekukonko/tablewriter"
)

func showResource(resourceRenderMap renderMap) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetRowLine(true)
	for k, v := range resourceRenderMap {
		table.Append([]string{strings.ToUpper(k), v})
	}
	table.Render()
}

func listResources(resourceRenderMaps []renderMap) {
	if len(resourceRenderMaps) > 0 {
		keys := []string{}
		for k := range resourceRenderMaps[0] {
			keys = append(keys, k)
		}
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader(keys)
		for _, m := range resourceRenderMaps {
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
