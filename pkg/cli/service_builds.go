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

func ShowServiceBuild(build *types.ServiceBuild, output OutputFormat) {
	switch output {
	case TABLE_OUTPUT:
		rm := getServiceBuildRenderMap(build)
		showResource(rm)
	case JSON_OUTPUT:
		DisplayAsJSON(build)
	}
}

func ShowServiceBuilds(builds []types.ServiceBuild, output OutputFormat) {
	switch output {
	case TABLE_OUTPUT:
		renderMaps := make([]renderMap, len(builds))
		for i, b := range builds {
			renderMaps[i] = getServiceBuildRenderMap(&b)
		}
		listResources(renderMaps)
	case JSON_OUTPUT:
		DisplayAsJSON(builds)
	}
}
