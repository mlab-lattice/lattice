package cli

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
	for _, v := range resources {
		ShowResource(v, asJSON)
	}
}
