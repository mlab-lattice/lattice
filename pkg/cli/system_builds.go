package cli

import (
	"github.com/mlab-lattice/system/pkg/types"
)

func getSystemBuildRenderMap(sb types.SystemBuild) map[string]string {
	return map[string]string{
		"ID":      string(sb.ID),
		"State":   string(sb.State),
		"Version": string(sb.Version),
	}
}

func ShowSystemBuild(build types.SystemBuild) {
	renderMap := getSystemBuildRenderMap(build)
	showResource(renderMap)
}

func ShowSystemBuilds(builds []types.SystemBuild) {
	renderMaps := make([]map[string]string, len(builds))
	for i, b := range builds {
		renderMaps[i] = getSystemBuildRenderMap(b)
	}
	listResources(renderMaps)
}
