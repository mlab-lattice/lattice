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

func ShowComponentBuild(build *types.ComponentBuild, output OutputFormat) {
	switch output {
	case TABLE_OUTPUT:
		rm := getComponentBuildRenderMap(build)
		showResource(rm)
	case JSON_OUTPUT:
		DisplayAsJSON(build)
	}
}

func ShowComponentBuilds(builds []types.ComponentBuild, output OutputFormat) {
	switch output {
	case TABLE_OUTPUT:
		renderMaps := make([]renderMap, len(builds))
		for i, b := range builds {
			renderMaps[i] = getComponentBuildRenderMap(&b)
		}
		listResources(renderMaps)
	case JSON_OUTPUT:
		DisplayAsJSON(builds)
	}
}

func ShowComponentBuildLog(stream io.Reader) {
	io.Copy(os.Stdout, stream)
}
