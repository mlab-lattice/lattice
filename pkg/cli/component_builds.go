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

func ShowComponentBuild(build *types.ComponentBuild) {
	rm := getComponentBuildRenderMap(build)
	showResource(rm)
}

func ShowComponentBuilds(builds []types.ComponentBuild) {
	renderMaps := make([]renderMap, len(builds))
	for i, b := range builds {
		renderMaps[i] = getComponentBuildRenderMap(&b)
	}
	listResources(renderMaps)
}

func ShowComponentBuildLog(stream io.Reader) {
	io.Copy(os.Stdout, stream)
}
