package componentbuilds

import (
	"io"
	"log"
	"os"

	"github.com/mlab-lattice/system/pkg/cli/resources"
	"github.com/mlab-lattice/system/pkg/types"
)

type ComponentBuildClient struct {
	resources.BuildClient
}

func (cbc ComponentBuildClient) List() {
	builds, err := cbc.RestClient.ComponentBuilds()
	if err != nil {
		log.Panic(err)
	}

	rs := []resources.EndpointResource{}
	for _, b := range builds {
		rs = append(rs, b)
	}
	resources.ListResources(rs, cbc.DisplayAsJSON)
}

func (cbc ComponentBuildClient) Show(id types.ComponentBuildID) {
	build, err := cbc.RestClient.ComponentBuild(id).Get()
	if err != nil {
		log.Panic(err)
	}
	resources.ShowResource(build, cbc.DisplayAsJSON)
}

func (cbc ComponentBuildClient) GetLogs(id types.ComponentBuildID, follow bool) {
	logs, err := cbc.RestClient.ComponentBuild(id).Logs(follow)
	if err != nil {
		log.Fatal(err)
	}
	defer logs.Close()

	io.Copy(os.Stdout, logs)
}
