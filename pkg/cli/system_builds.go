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

func ShowSystemBuild(build *types.SystemBuild, output OutputFormat) error {
	switch output {
	case OutputFormatTable:
		rm := getSystemBuildRenderMap(build)
		showResource(rm)
	case OutputFormatJSON:
		DisplayAsJSON(build)
	default:
		return newOutputFormatError(output)
	}
	return nil
}

func ShowSystemBuilds(builds []types.SystemBuild, output OutputFormat) error {
	switch output {
	case OutputFormatTable:
		renderMaps := make([]renderMap, len(builds))
		for i, b := range builds {
			renderMaps[i] = getSystemBuildRenderMap(&b)
		}
		listResources(renderMaps)
	case OutputFormatJSON:
		DisplayAsJSON(builds)
	default:
		return newOutputFormatError(output)
	}
	return nil
}
