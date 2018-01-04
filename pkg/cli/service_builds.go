package cli

import (
	"github.com/mlab-lattice/system/pkg/types"
)

func getServiceBuildRenderMap(build *types.ServiceBuild) RenderMap {
	return RenderMap{
		"ID":    string(build.ID),
		"State": string(build.State),
	}
}

func ShowServiceBuild(build *types.ServiceBuild, output OutputFormat) error {
	switch output {
	case OutputFormatTable:
		rm := getServiceBuildRenderMap(build)
		ShowResource(rm)
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
		renderMaps := make([]RenderMap, len(builds))
		for i, b := range builds {
			renderMaps[i] = getServiceBuildRenderMap(&b)
		}
		ListResources(renderMaps)
	case OutputFormatJSON:
		DisplayAsJSON(builds)
	default:
		return NewOutputFormatError(output)
	}
	return nil
}
