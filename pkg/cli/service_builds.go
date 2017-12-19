package cli

import (
	"github.com/mlab-lattice/system/pkg/types"
)

func getServiceBuildRenderMap(sb types.ServiceBuild) map[string]string {
	return map[string]string{
		"ID":    string(sb.ID),
		"State": string(sb.State),
	}
}

func ShowServiceBuild(build types.ServiceBuild) {
	renderMap := getServiceBuildRenderMap(build)
	showResource(renderMap)
}

func ShowServiceBuilds(builds []types.ServiceBuild) {
	renderMaps := make([]map[string]string, len(builds))
	for i, b := range builds {
		renderMaps[i] = getServiceBuildRenderMap(b)
	}
	listResources(renderMaps)
}
