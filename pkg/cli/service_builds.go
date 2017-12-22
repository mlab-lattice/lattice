package cli

import (
	"github.com/mlab-lattice/system/pkg/types"
)

func getServiceBuildRenderMap(build *types.ServiceBuild) renderMap {
	return renderMap{
		"ID":    string(build.ID),
		"State": string(build.State),
	}
}

func ShowServiceBuild(build *types.ServiceBuild, output OutputFormat) error {
	switch output {
	case OutputFormatTable:
		rm := getServiceBuildRenderMap(build)
		showResource(rm)
	case OutputFormatJSON:
		DisplayAsJSON(build)
	default:
		return newOutputFormatError(output)
	}
	return nil
}

func ShowServiceBuilds(builds []types.ServiceBuild, output OutputFormat) error {
	switch output {
	case OutputFormatTable:
		renderMaps := make([]renderMap, len(builds))
		for i, b := range builds {
			renderMaps[i] = getServiceBuildRenderMap(&b)
		}
		listResources(renderMaps)
	case OutputFormatJSON:
		DisplayAsJSON(builds)
	default:
		return newOutputFormatError(output)
	}
	return nil
}
