package e2e

import (
	"flag"
	"testing"

	// test sources
	"github.com/mlab-lattice/lattice/test/e2e/context"
	_ "github.com/mlab-lattice/lattice/test/e2e/system"

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
	context.SetClusterURL(clusterURL)
	provisioned = true

	return nil
}, func([]byte) {})

func init() {
	flag.StringVar(&clusterURL, "cluster-url", "", "cluster url")

	flag.StringVar(&provider, "cloud-provider", "", "cloud provider")
	flag.StringVar(&controlPlaneContainerRegistry, "control-plane-container-registry", "gcr.io/lattice-dev", "container registry for control plane containers")
	flag.StringVar(&controlPlaneContainerChannel, "control-plane-container-channel", "stable-debug-", "channel of control plane containers")
}
