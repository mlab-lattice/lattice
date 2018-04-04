package e2e

import (
	"flag"
	"testing"

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

func init() {
	flag.StringVar(&clusterURL, "cluster-url", "", "cluster url")

	flag.StringVar(&provider, "cloud-provider", "", "cloud provider")
	flag.StringVar(&controlPlaneContainerRegistry, "control-plane-container-registry", "gcr.io/lattice-dev", "container registry for control plane containers")
	flag.StringVar(&controlPlaneContainerChannel, "control-plane-container-channel", "stable-debug-", "channel of control plane containers")
}
