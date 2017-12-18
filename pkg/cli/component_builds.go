package cli

import (
	"io"
	"os"

	"github.com/mlab-lattice/system/pkg/cli/resources"
	"github.com/mlab-lattice/system/pkg/types"
)

func ShowComponentBuild(build types.ComponentBuild) {
	showResource(build)
}

func ShowComponentBuilds(builds []types.ComponentBuild) {
	rs := []resources.EndpointResource{}
	for _, b := range builds {
		rs = append(rs, b)
	}
	listResources(rs)
}

func ShowComponentBuildLog(stream io.ReadCloser) {
	defer stream.Close()
	io.Copy(os.Stdout, stream)
}
