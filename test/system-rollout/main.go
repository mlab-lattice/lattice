package main

import (
	"flag"
	"fmt"

	"github.com/mlab-lattice/core/pkg/constants"
	coretypes "github.com/mlab-lattice/core/pkg/types"

	"github.com/mlab-lattice/kubernetes-integration/pkg/system-environment-manager/backend"
)

var (
	kubeconfig string
	buildId    string
)

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "path to kubeconfig file")
	flag.StringVar(&buildId, "buildId", "", "id of the SystemBuild to roll out")
	flag.Parse()

}

func main() {
	kb, err := backend.NewKubernetesBackend(kubeconfig)
	if err != nil {
		panic(err)
	}

	sysRolloutId, err := kb.RollOutSystemBuild(constants.UserSystemNamespace, coretypes.SystemBuildId(buildId))
	if err != nil {
		panic(err)
	}

	fmt.Printf("Created SystemRollout %v\n", sysRolloutId)
}
