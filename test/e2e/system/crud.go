package system

import (
	"time"

	"github.com/mlab-lattice/system/pkg/types"
	"github.com/mlab-lattice/system/test/e2e/context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("system", func() {
	It("should be able to list systems, but the list should be empty", func() {
		//systems, err := client.Systems().List()
		systems, err := context.TestContext.ClusterAPIClient.Systems().List()
		Expect(err).NotTo(HaveOccurred())

		Expect(len(systems)).To(Equal(0))
	})

	systemID := types.SystemID("e2e-system-1")
	systemURL := "https://github.com/mlab-lattice/testing__system.git"
	It("should be able to create a system", func() {
		system, err := context.TestContext.ClusterAPIClient.Systems().Create(systemID, systemURL)
		Expect(err).NotTo(HaveOccurred())

		Expect(system).To(Not(Equal(nil)))

		Expect(system.ID).To(Equal(systemID))
		Expect(system.DefinitionURL).To(Equal(systemURL))
		Expect(len(system.Services)).To(Equal(0))
		Expect(system.State).To(Equal(types.SystemStateStable))
	})

	It("should be able to list systems, and there should only be the newly created system", func() {
		systems, err := context.TestContext.ClusterAPIClient.Systems().List()
		Expect(err).NotTo(HaveOccurred())

		Expect(len(systems)).To(Equal(1))

		system := systems[0]
		Expect(system.ID).To(Equal(systemID))
		Expect(system.DefinitionURL).To(Equal(systemURL))
		Expect(len(system.Services)).To(Equal(0))
		Expect(system.State).To(Equal(types.SystemStateStable))
	})

	It("should be able to get the newly created system by ID", func() {
		system, err := context.TestContext.ClusterAPIClient.Systems().Get(systemID)
		Expect(err).NotTo(HaveOccurred())

		Expect(system).To(Not(Equal(nil)))

		Expect(system.ID).To(Equal(systemID))
		Expect(system.DefinitionURL).To(Equal(systemURL))
		Expect(len(system.Services)).To(Equal(0))
		Expect(system.State).To(Equal(types.SystemStateStable))
	})

	// Wait to ensure controller sees the system and updates the status
	time.Sleep(1 * time.Second)
	It("should be able to delete the newly created system by ID", func() {
		err := context.TestContext.ClusterAPIClient.Systems().Delete(types.SystemID(systemID))
		Expect(err).NotTo(HaveOccurred())
	})

	It("should be able to list systems, but the list should be empty", func() {
		systems, err := context.TestContext.ClusterAPIClient.Systems().List()
		Expect(err).NotTo(HaveOccurred())

		Expect(len(systems)).To(Equal(0))
	})
})
