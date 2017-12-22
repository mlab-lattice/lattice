package cli

import (
	"io"
	"os"

	"github.com/mlab-lattice/system/pkg/types"
)

func getComponentBuildRenderMap(build *types.ComponentBuild) renderMap {
	return renderMap{
		"ID":    string(build.ID),
		"State": string(build.State),
	}
}

func ShowComponentBuild(build *types.ComponentBuild, output OutputFormat) error {
	switch output {
	case OutputFormatTable:
		rm := getComponentBuildRenderMap(build)
		showResource(rm)
	case OutputFormatJSON:
		DisplayAsJSON(build)
	default:
		return newOutputFormatError(output)
	}
	return nil
}

func ShowComponentBuilds(builds []types.ComponentBuild, output OutputFormat) error {
	switch output {
	case OutputFormatTable:
		renderMaps := make([]renderMap, len(builds))
		for i, b := range builds {
			renderMaps[i] = getComponentBuildRenderMap(&b)
		}
		listResources(renderMaps)
	case OutputFormatJSON:
		DisplayAsJSON(builds)
	default:
		return newOutputFormatError(output)
	}
	return nil
}

func ShowComponentBuildLog(stream io.Reader) {
	io.Copy(os.Stdout, stream)
}
