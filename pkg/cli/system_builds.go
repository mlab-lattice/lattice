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

func ShowSystemBuild(build *types.SystemBuild, output OutputFormat) {
	switch output {
	case TABLE_OUTPUT:
		rm := getSystemBuildRenderMap(build)
		showResource(rm)
	case JSON_OUTPUT:
		DisplayAsJSON(build)
	}
}

func ShowSystemBuilds(builds []types.SystemBuild, output OutputFormat) {
	switch output {
	case TABLE_OUTPUT:
		renderMaps := make([]renderMap, len(builds))
		for i, b := range builds {
			renderMaps[i] = getSystemBuildRenderMap(&b)
		}
		listResources(renderMaps)
	case JSON_OUTPUT:
		DisplayAsJSON(builds)
	}
}
