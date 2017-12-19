package cli

import (
	"github.com/mlab-lattice/system/pkg/types"
)

func getSystemBuildRenderMap(build *types.SystemBuild) RenderMap {
	return RenderMap{
		"ID":      string(build.ID),
		"State":   string(build.State),
		"Version": string(build.Version),
	}
}

func ShowSystemBuild(build *types.SystemBuild) {
	renderMap := getSystemBuildRenderMap(build)
	showResource(renderMap)
}

func ShowSystemBuilds(builds []types.SystemBuild) {
	renderMaps := make([]RenderMap, len(builds))
	for i, b := range builds {
		renderMaps[i] = getSystemBuildRenderMap(&b)
	}
	listResources(renderMaps)
}
