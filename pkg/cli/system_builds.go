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

func ShowSystemBuild(build *types.SystemBuild, output OutputFormat) error {
	switch output {
	case OutputFormatTable:
		rm := getSystemBuildRenderMap(build)
		ShowResource(rm)
	case OutputFormatJSON:
		DisplayAsJSON(build)
	default:
		return NewOutputFormatError(output)
	}
	return nil
}

func ShowSystemBuilds(builds []types.SystemBuild, output OutputFormat) error {
	switch output {
	case OutputFormatTable:
		renderMaps := make([]RenderMap, len(builds))
		for i, b := range builds {
			renderMaps[i] = getSystemBuildRenderMap(&b)
		}
		ListResources(renderMaps)
	case OutputFormatJSON:
		DisplayAsJSON(builds)
	default:
		return NewOutputFormatError(output)
	}
	return nil
}
