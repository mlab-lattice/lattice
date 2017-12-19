package cli

import (
	"github.com/mlab-lattice/system/pkg/types"
)

func getSystemBuildRenderMap(build *types.SystemBuild) renderMap {
	return renderMap{
		"ID":      string(build.ID),
		"State":   string(build.State),
		"Version": string(build.Version),
	}
}

func ShowSystemBuild(build *types.SystemBuild) {
	rm := getSystemBuildRenderMap(build)
	showResource(rm)
}

func ShowSystemBuilds(builds []types.SystemBuild) {
	renderMaps := make([]renderMap, len(builds))
	for i, b := range builds {
		renderMaps[i] = getSystemBuildRenderMap(&b)
	}
	listResources(renderMaps)
}
