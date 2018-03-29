package e2e

import (
	"flag"
	"testing"

	"github.com/mlab-lattice/system/pkg/backend/kubernetes/cloudprovider"
	"github.com/mlab-lattice/system/pkg/backend/kubernetes/cloudprovider/local"
	"github.com/mlab-lattice/system/pkg/lifecycle/lattice/provisioner"
	"github.com/mlab-lattice/system/test/e2e/context"

	// test sources
	_ "github.com/mlab-lattice/system/test/e2e/system"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
)

var (
	provider                      string
	controlPlaneContainerRegistry string
	controlPlaneContainerChannel  string

	clusterURL string

	provisioned bool
)

func RunE2ETest(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Lattice e2e Suite")
}

var _ = ginkgo.SynchronizedBeforeSuite(func() []byte {
	if clusterURL == "" {
		p := getProvisioner()

		var err error
		clusterURL, err = p.Provision("e2e-test")
		if err != nil {
			panic(err)
		}

	}

	context.SetClusterURL(clusterURL)
	provisioned = true

	return nil
}, func([]byte) {
})

var _ = ginkgo.SynchronizedAfterSuite(func() {
	if provider != "" {
		p := getProvisioner()

		err := p.Deprovision("e2e-test", !provisioned)
		if err != nil {
			panic(err)
		}
	}
}, func() {

})

func getProvisioner() provisioner.Interface {
	switch provider {
	case cloudprovider.Local:
		p, err := local.NewLatticeProvisioner(
			controlPlaneContainerRegistry,
			controlPlaneContainerChannel,
			"/tmp/lattice/test/e2e/local",
			&local.LatticeProvisionerOptions{},
		)
		if err != nil {
			panic(err)
		}

		return p

	default:
		panic("unsupported provider " + provider)
	}
}

func init() {
	flag.StringVar(&clusterURL, "cluster-url", "", "cluster url")

	flag.StringVar(&provider, "cloud-provider", "", "cloud provider")
	flag.StringVar(&controlPlaneContainerRegistry, "control-plane-container-registry", "gcr.io/lattice-dev", "container registry for control plane containers")
	flag.StringVar(&controlPlaneContainerChannel, "control-plane-container-channel", "stable-debug-", "channel of control plane containers")
}
