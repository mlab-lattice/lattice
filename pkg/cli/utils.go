package cli

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/olekukonko/tablewriter"
)

func showResource(renderMap map[string]string) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetRowLine(true)
	for k, v := range renderMap {
		table.Append([]string{strings.ToUpper(k), v})
	}
	table.Render()
}

func listResources(renderMaps []map[string]string) {
	if len(renderMaps) > 0 {
		keys := []string{}
		for k := range renderMaps[0] {
			keys = append(keys, k)
		}
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader(keys)
		for _, m := range renderMaps {
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
