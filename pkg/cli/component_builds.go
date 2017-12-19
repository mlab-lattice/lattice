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

func ShowComponentBuild(build *types.ComponentBuild) {
	renderMap := getComponentBuildRenderMap(build)
	showResource(renderMap)
}

func ShowComponentBuilds(builds []types.ComponentBuild) {
	renderMaps := make([]RenderMap, len(builds))
	for i, b := range builds {
		renderMaps[i] = getComponentBuildRenderMap(&b)
	}
	listResources(renderMaps)
}

func ShowComponentBuildLog(stream io.Reader) {
	io.Copy(os.Stdout, stream)
}
