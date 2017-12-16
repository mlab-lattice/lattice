package systembuilds

import (
	"fmt"
	"log"

	"github.com/mlab-lattice/system/pkg/cli/resources"
	"github.com/mlab-lattice/system/pkg/types"
)

type SystemBuildClient struct {
	resources.BuildClient
}

func (sbc SystemBuildClient) List() {
	builds, err := sbc.RestClient.SystemBuilds()
	fmt.Println(builds)
	if err != nil {
		log.Panic(err)
	}

	rs := []resources.EndpointResource{}
	for _, b := range builds {
		rs = append(rs, b)
	}
	resources.ListResources(rs, sbc.DisplayAsJSON)
}

func (sbc SystemBuildClient) Show(id types.SystemBuildID) {
	build, err := sbc.RestClient.SystemBuild(id).Get()
	if err != nil {
		log.Panic(err)
	}
	resources.ShowResource(build, sbc.DisplayAsJSON)
}
