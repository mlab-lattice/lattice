package cli

import (
	"github.com/mlab-lattice/system/pkg/types"
)

var teardownHeaders = []string{"ID", "State"}

func getTeardownValues(teardown *types.SystemTeardown) []string {
	return []string{
		string(teardown.ID),
		string(teardown.State),
	}
}

func ShowTeardown(teardown *types.SystemTeardown, output OutputFormat) error {
	switch output {
	case OutputFormatTable:
		values := getTeardownValues(teardown)
		ShowResource(teardownHeaders, values)
	case OutputFormatJSON:
		DisplayAsJSON(teardown)
	default:
		return NewOutputFormatError(output)
	}
	return nil
}

func ShowTeardowns(teardowns []types.SystemTeardown, output OutputFormat) error {
	switch output {
	case OutputFormatTable:
		values := make([][]string, len(teardowns))
		for i, b := range teardowns {
			values[i] = getTeardownValues(&b)
		}
		ListResources(teardownHeaders, values)
	case OutputFormatJSON:
		DisplayAsJSON(teardowns)
	default:
		return NewOutputFormatError(output)
	}
	return nil
}
