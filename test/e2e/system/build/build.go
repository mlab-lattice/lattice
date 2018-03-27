package build

import (
	"time"

	"github.com/mlab-lattice/system/pkg/api/v1"
	"github.com/mlab-lattice/system/test/e2e/context"

	"k8s.io/apimachinery/pkg/util/wait"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("build", func() {
	systemID := v1.SystemID("e2e-system-build-1")
	systemURL := "https://github.com/mlab-lattice/testing__system.git"

	It("should be able to create a system", func() {
		system, err := context.TestContext.LatticeAPIClient.Systems().Create(systemID, systemURL)
		Expect(err).NotTo(HaveOccurred())

		Expect(system).To(Not(Equal(nil)))

		Expect(system.ID).To(Equal(systemID))
		Expect(system.DefinitionURL).To(Equal(systemURL))
		Expect(len(system.Services)).To(Equal(0))
		Expect(system.State).To(Equal(v1.SystemStatePending))

		Eventually(func() v1.SystemState {
			system, err := context.TestContext.LatticeAPIClient.Systems().Get(systemID)
			if err != nil {
				return v1.SystemStateFailed
			}
			return system.State
		}, 10*time.Second).Should(Equal(v1.SystemStateStable))
	})

	It("should be able to list builds, but the list should be empty", func() {
		systemBuilds, err := context.TestContext.LatticeAPIClient.Systems().Builds(systemID).List()
		Expect(err).NotTo(HaveOccurred())

		Expect(len(systemBuilds)).To(Equal(0))
	})

	version := v1.SystemVersion("v3.0.0")
	var build1ID v1.BuildID
	It("should be able to create a build", func() {
		build, err := context.TestContext.LatticeAPIClient.Systems().Builds(systemID).Create(version)
		Expect(err).NotTo(HaveOccurred())

		build1ID = build.ID
		Expect(build.Version).To(Equal(version))
		Expect(build.State).To(Equal(v1.BuildStatePending))
	})

	It("should be able to get the build by ID", func() {
		build, err := context.TestContext.LatticeAPIClient.Systems().Builds(systemID).Get(build1ID)
		Expect(err).NotTo(HaveOccurred())

		Expect(build.Version).To(Equal(version))
		Expect(build.State).To(SatisfyAny(Equal(v1.BuildStatePending), Equal(v1.BuildStateRunning)))
	})

	It("should be able to list builds, and the list should only contain the created build", func() {
		builds, err := context.TestContext.LatticeAPIClient.Systems().Builds(systemID).List()
		Expect(err).NotTo(HaveOccurred())

		Expect(len(builds)).To(Equal(1))

		build := builds[0]
		Expect(build.Version).To(Equal(version))
		Expect(build.State).To(SatisfyAny(Equal(v1.BuildStatePending), Equal(v1.BuildStateRunning)))
	})

	It("should see the build succeed", func() {
		Eventually(func() v1.BuildState {
			build, err := context.TestContext.LatticeAPIClient.Systems().Builds(systemID).Get(build1ID)
			if err != nil {
				return v1.BuildStateFailed
			}
			return build.State
		}, 10*time.Minute).Should(Equal(v1.BuildStateSucceeded))
	})

	var build2ID v1.BuildID
	It("should be able to build the same version again, much faster", func() {
		build, err := context.TestContext.LatticeAPIClient.Systems().Builds(systemID).Create(version)
		Expect(err).NotTo(HaveOccurred())

		build2ID = build.ID
		Expect(build.ID).NotTo(Equal(build1ID))
		Expect(build.Version).To(Equal(version))
		Expect(build.State).To(Equal(v1.BuildStatePending))

		Eventually(func() v1.BuildState {
			build, err := context.TestContext.LatticeAPIClient.Systems().Builds(systemID).Get(build1ID)
			if err != nil {
				return v1.BuildStateFailed
			}
			return build.State
		}, 30*time.Second).Should(Equal(v1.BuildStateSucceeded))
	})

	It("should be able to list builds, and the list should contain both builds", func() {
		builds, err := context.TestContext.LatticeAPIClient.Systems().Builds(systemID).List()
		Expect(err).NotTo(HaveOccurred())

		Expect(len(builds)).To(Equal(2))

		seenBuilds := map[v1.BuildID]struct{}{}
		for _, build := range builds {
			_, ok := seenBuilds[build.ID]
			Expect(ok).To(Equal(false))
			Expect(build.Version).To(Equal(version))
			Expect(build.State).To(Equal(v1.BuildStateSucceeded))
			seenBuilds[build.ID] = struct{}{}
		}
	})

	It("should be able to delete the system", func() {
		err := context.TestContext.LatticeAPIClient.Systems().Delete(v1.SystemID(systemID))
		Expect(err).NotTo(HaveOccurred())

		err = wait.PollImmediate(time.Second, 45*time.Second, func() (bool, error) {
			_, err := context.TestContext.LatticeAPIClient.Systems().Get(systemID)
			if err != nil {
				if _, ok := err.(*v1.InvalidSystemIDError); ok {
					return true, nil
				}

				return false, err
			}

			return false, nil
		})
		Expect(err).NotTo(HaveOccurred())
	})
})
