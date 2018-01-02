package cli

import (
	"github.com/mlab-lattice/system/pkg/types"
)

var serviceBuildHeaders = []string{"ID", "State"}

func getServiceBuildValues(build *types.ServiceBuild) []string {
	return []string{
		string(build.ID),
		string(build.State),
	}
}

func ShowServiceBuild(build *types.ServiceBuild, output OutputFormat) error {
	switch output {
	case OutputFormatTable:
		values := getServiceBuildValues(build)
		ShowResource(serviceBuildHeaders, values)
	case OutputFormatJSON:
		DisplayAsJSON(build)
	default:
		return NewOutputFormatError(output)
	}
	return nil
}

func ShowServiceBuilds(builds []types.ServiceBuild, output OutputFormat) error {
	switch output {
	case OutputFormatTable:
		values := make([][]string, len(builds))
		for i, b := range builds {
			values[i] = getServiceBuildValues(&b)
		}
		ListResources(serviceBuildHeaders, values)
	case OutputFormatJSON:
		DisplayAsJSON(builds)
	default:
		return NewOutputFormatError(output)
	}
	return nil
}
