package resources

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/olekukonko/tablewriter"
)

type EndpointResource interface {
	GetRenderMap() map[string]string
}

func ShowResource(resource EndpointResource, asJSON bool) {
	if asJSON {
		buf, err := json.MarshalIndent(resource, "", "  ")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(buf))
	} else {
		renderMap := resource.GetRenderMap()
		table := tablewriter.NewWriter(os.Stdout)
		table.SetRowLine(true)
		for k, v := range renderMap {
			table.Append([]string{strings.ToUpper(k), v})
		}
		table.Render()
	}
}

func ListResources(resources []EndpointResource, asJSON bool) {
	if asJSON {
		buf, err := json.MarshalIndent(resources, "", "  ")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(buf))
	} else if len(resources) > 0 {
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
