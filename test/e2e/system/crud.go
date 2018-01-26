package system

import (
	"github.com/mlab-lattice/system/test/e2e/context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("system", func() {
	It("should be able to list systems, but the list should be empty", func() {
		systems, err := context.TestContext.ClusterAPIClient.Systems().List()
		Expect(err).NotTo(HaveOccurred())

		Expect(len(systems)).To(Equal(0))
	})
})
