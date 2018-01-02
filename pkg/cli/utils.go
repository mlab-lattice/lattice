package cli

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/olekukonko/tablewriter"
)

func ShowResource(headers []string, values []string) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetRowLine(true)
	table.SetHeader(headers)
	for i, h := range headers {
		table.Append([]string{strings.ToUpper(h), values[i]})
	}
	table.Render()
}

func ListResources(headers []string, values [][]string) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(headers)
	for _, line := range values {
		table.Append(line)
	}
	table.Render()
}

func DisplayAsJSON(resource interface{}) {
	buf, err := json.MarshalIndent(resource, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(buf))
}
