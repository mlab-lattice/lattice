package cli

import (
	"github.com/mlab-lattice/system/pkg/cli/resources"
	"github.com/mlab-lattice/system/pkg/types"
)

func ShowServiceBuild(build types.ServiceBuild) {
	showResource(build)
}

func ShowServiceBuilds(builds []types.ServiceBuild) {
	rs := []resources.EndpointResource{}
	for _, b := range builds {
		rs = append(rs, b)
	}
	listResources(rs)
}
