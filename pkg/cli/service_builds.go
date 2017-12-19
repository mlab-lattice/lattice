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

func ShowServiceBuild(build *types.ServiceBuild) {
	rm := getServiceBuildRenderMap(build)
	showResource(rm)
}

func ShowServiceBuilds(builds []types.ServiceBuild) {
	renderMaps := make([]renderMap, len(builds))
	for i, b := range builds {
		renderMaps[i] = getServiceBuildRenderMap(&b)
	}
	listResources(renderMaps)
}
