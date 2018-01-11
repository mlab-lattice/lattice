package cli

import (
	"fmt"
	"strings"

	"github.com/mlab-lattice/system/pkg/types"
)

var systemHeaders = []string{
	"ID",
	"State",
	"Definition URL",
	"Services",
}

func getSystemValues(system *types.System) []string {
	values := []string{
		string(system.ID),
		string(system.State),
		string(system.DefinitionURL),
	}

	services := ""
	for path, service := range system.Services {
		services += fmt.Sprintf("%v: %v\n", path, service.ID)
	}

	services = strings.TrimSpace(services)
	values = append(values, services)

	return values
}

func ShowSystem(system *types.System, output OutputFormat) error {
	switch output {
	case OutputFormatTable:
		values := getSystemValues(system)
		ShowResource(systemHeaders, values)
	case OutputFormatJSON:
		DisplayAsJSON(system)
	default:
		return NewOutputFormatError(output)
	}
	return nil
}

func ShowSystems(systems []types.System, output OutputFormat) error {
	switch output {
	case OutputFormatTable:
		values := make([][]string, len(systems))
		for i, b := range systems {
			values[i] = getSystemValues(&b)
		}
		ListResources(systemHeaders, values)
	case OutputFormatJSON:
		DisplayAsJSON(systems)
	default:
		return NewOutputFormatError(output)
	}
	return nil
}
