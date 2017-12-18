package cli

import (
	"github.com/mlab-lattice/system/pkg/cli/resources"
	"github.com/mlab-lattice/system/pkg/types"
)

func ShowSystemBuild(build types.SystemBuild) {
	showResource(build)
}

func ShowSystemBuilds(builds []types.SystemBuild) {
	rs := []resources.EndpointResource{}
	for _, b := range builds {
		rs = append(rs, b)
	}
	listResources(rs)
}
