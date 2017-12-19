package cli

import (
	"io"
	"os"

	"github.com/mlab-lattice/system/pkg/types"
)

func getComponentBuildRenderMap(sb types.ComponentBuild) map[string]string {
	return map[string]string{
		"ID":    string(sb.ID),
		"State": string(sb.State),
	}
}

func ShowComponentBuild(build types.ComponentBuild) {
	renderMap := getComponentBuildRenderMap(build)
	showResource(renderMap)
}

func ShowComponentBuilds(builds []types.ComponentBuild) {
	renderMaps := make([]map[string]string, len(builds))
	for i, b := range builds {
		renderMaps[i] = getComponentBuildRenderMap(b)
	}
	listResources(renderMaps)
}

func ShowComponentBuildLog(stream io.ReadCloser) {
	defer stream.Close()
	io.Copy(os.Stdout, stream)
}
