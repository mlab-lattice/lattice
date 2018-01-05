package cli

import (
	"io"
	"os"

	"github.com/mlab-lattice/system/pkg/types"
)

var componentBuildHeaders = []string{"ID", "State"}

func getComponentBuildValues(build *types.ComponentBuild) []string {
	return []string{
		string(build.ID),
		string(build.State),
	}
}

func ShowComponentBuild(build *types.ComponentBuild, output OutputFormat) error {
	switch output {
	case OutputFormatTable:
		values := getComponentBuildValues(build)
		ShowResource(componentBuildHeaders, values)
	case OutputFormatJSON:
		DisplayAsJSON(build)
	default:
		return NewOutputFormatError(output)
	}
	return nil
}

func ShowComponentBuilds(builds []types.ComponentBuild, output OutputFormat) error {
	switch output {
	case OutputFormatTable:
		values := make([][]string, len(builds))
		for i, b := range builds {
			values[i] = getComponentBuildValues(&b)
		}
		ListResources(componentBuildHeaders, values)
	case OutputFormatJSON:
		DisplayAsJSON(builds)
	default:
		return NewOutputFormatError(output)
	}
	return nil
}

func ShowComponentBuildLog(stream io.Reader) {
	io.Copy(os.Stdout, stream)
}
