package servicebuilds

import (
	"log"

	"github.com/mlab-lattice/system/pkg/cli/resources"
	"github.com/mlab-lattice/system/pkg/types"
)

type ServiceBuildClient struct {
	resources.BuildClient
}

func (sbc ServiceBuildClient) List() {
	builds, err := sbc.RestClient.ServiceBuilds()
	if err != nil {
		log.Panic(err)
	}

	rs := []resources.EndpointResource{}
	for _, b := range builds {
		rs = append(rs, b)
	}
	resources.ListResources(rs, sbc.DisplayAsJSON)
}

func (sbc ServiceBuildClient) Show(id types.ServiceBuildID) {
	build, err := sbc.RestClient.ServiceBuild(id).Get()
	if err != nil {
		log.Panic(err)
	}
	resources.ShowResource(build, sbc.DisplayAsJSON)
}
