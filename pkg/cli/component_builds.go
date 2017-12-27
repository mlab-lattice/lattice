package cli

import (
	"io"
	"os"

	"github.com/mlab-lattice/system/pkg/types"
)

func getComponentBuildRenderMap(build *types.ComponentBuild) RenderMap {
	return RenderMap{
		"ID":    string(build.ID),
		"State": string(build.State),
	}
}

func ShowComponentBuild(build *types.ComponentBuild, output OutputFormat) error {
	switch output {
	case OutputFormatTable:
		rm := getComponentBuildRenderMap(build)
		ShowResource(rm)
	case OutputFormatJSON:
		DisplayAsJSON(build)
	default:
		return NewOutputFormatError(output)
	}
	return nil
}

func ShowComponentBuilds(builds []types.ComponentBuild, output OutputFormat) error {
	switch output {
	case OutputFormatTable:
		renderMaps := make([]RenderMap, len(builds))
		for i, b := range builds {
			renderMaps[i] = getComponentBuildRenderMap(&b)
		}
		ListResources(renderMaps)
	case OutputFormatJSON:
		DisplayAsJSON(builds)
	default:
		return NewOutputFormatError(output)
	}
	return nil
}

func ShowComponentBuildLog(stream io.Reader) {
	io.Copy(os.Stdout, stream)
}
