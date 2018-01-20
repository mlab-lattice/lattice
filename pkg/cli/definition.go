package cli

import (
	"os"

	"github.com/olekukonko/tablewriter"
)

func ListSystemDefinitionVersions(versions []string) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetRowLine(true)
	table.SetAlignment(tablewriter.ALIGN_CENTER)

	for _, version := range versions {
		table.Append([]string{version})
	}

	table.Render()
}
