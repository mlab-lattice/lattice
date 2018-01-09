package cli

import (
	"github.com/mlab-lattice/system/pkg/types"
)

var systemBuildHeaders = []string{"ID", "State", "Version"}

func getSystemBuildValues(build *types.SystemBuild) []string {
	return []string{
		string(build.ID),
		string(build.State),
		string(build.Version),
	}
}

func ShowSystemBuild(build *types.SystemBuild, output OutputFormat) error {
	switch output {
	case OutputFormatTable:
		values := getSystemBuildValues(build)
		ShowResource(systemBuildHeaders, values)
	case OutputFormatJSON:
		DisplayAsJSON(build)
	default:
		return NewOutputFormatError(output)
	}
	return nil
}

func ShowSystemBuilds(builds []types.SystemBuild, output OutputFormat) error {
	switch output {
	case OutputFormatTable:
		values := make([][]string, len(builds))
		for i, b := range builds {
			values[i] = getSystemBuildValues(&b)
		}
		ListResources(componentBuildHeaders, values)
	case OutputFormatJSON:
		DisplayAsJSON(builds)
	default:
		return NewOutputFormatError(output)
	}
	return nil
}
