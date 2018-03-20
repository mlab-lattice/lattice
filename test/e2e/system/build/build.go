package build

//
//import (
//	"fmt"
//	//"time"
//
//	"github.com/mlab-lattice/system/pkg/types"
//	"github.com/mlab-lattice/system/test/e2e/context"
//
//	//"k8s.io/apimachinery/pkg/util/wait"
//
//	. "github.com/onsi/ginkgo"
//	. "github.com/onsi/gomega"
//)
//
//const (
//	systemID  = types.SystemID("e2e-system-build-1")
//	systemURL = "https://github.com/mlab-lattice/testing__system.git"
//)
//
//var _ = Describe("build", func() {
//	//_, err := context.TestContext.ClusterAPIClient.Systems().Create(systemID, systemURL)
//	//Expect(err).NotTo(HaveOccurred())
//
//	It("should be able to list builds, but the list should be empty", func() {
//		systemBuilds, err := context.TestContext.ClusterAPIClient.Systems().SystemBuilds(systemID).List()
//		Expect(err).NotTo(HaveOccurred())
//
//		Expect(len(systemBuilds)).To(Equal(0))
//	})
//
//	It("should be able to list builds, but the list should be empty", func() {
//		systemBuilds, err := context.TestContext.ClusterAPIClient.Systems().SystemBuilds(systemID).List()
//		Expect(err).NotTo(HaveOccurred())
//
//		Expect(len(systemBuilds)).To(Equal(0))
//	})
//
//	//err = context.TestContext.ClusterAPIClient.Systems().Delete(types.SystemID(systemID))
//	//Expect(err).NotTo(HaveOccurred())
//	//
//	//err = wait.PollImmediate(time.Second, 15*time.Second, func() (bool, error) {
//	//	_, err := context.TestContext.ClusterAPIClient.Systems().Get(systemID)
//	//	if err != nil {
//	//		 // FIXME: no way to tell between DNE and normal error right now
//	//return true, nil
//	//}
//	//
//	//return false, nil
//	//})
//	//Expect(err).NotTo(HaveOccurred())
//
//	//It("should be able to list systems, and there should only be the newly created system", func() {
//	//	systems, err := context.TestContext.ClusterAPIClient.Systems().List()
//	//	Expect(err).NotTo(HaveOccurred())
//	//
//	//	Expect(len(systems)).To(Equal(1))
//	//
//	//	system := systems[0]
//	//	Expect(system.ID).To(Equal(systemID))
//	//	Expect(system.DefinitionURL).To(Equal(systemURL))
//	//	Expect(len(system.Services)).To(Equal(0))
//	//	Expect(system.State).To(Equal(types.SystemStateStable))
//	//})
//	//
//	//It("should be able to get the newly created system by ID", func() {
//	//	system, err := context.TestContext.ClusterAPIClient.Systems().Get(systemID)
//	//	Expect(err).NotTo(HaveOccurred())
//	//
//	//	Expect(system).To(Not(Equal(nil)))
//	//
//	//	Expect(system.ID).To(Equal(systemID))
//	//	Expect(system.DefinitionURL).To(Equal(systemURL))
//	//	Expect(len(system.Services)).To(Equal(0))
//	//	Expect(system.State).To(Equal(types.SystemStateStable))
//	//})
//	//
//	//// Wait to ensure controller sees the system and updates the status
//	//time.Sleep(3 * time.Second)
//	//It("should be able to delete the newly created system by ID", func() {
//	//	err := context.TestContext.ClusterAPIClient.Systems().Delete(types.SystemID(systemID))
//	//	Expect(err).NotTo(HaveOccurred())
//	//})
//	//
//	//It("should be able to list systems, and the deleted system should either be in the deleting state or no longer be in the list", func() {
//	//	err := wait.PollImmediate(time.Second, 15*time.Second, func() (bool, error) {
//	//		systems, err := context.TestContext.ClusterAPIClient.Systems().List()
//	//		if err != nil {
//	//			return false, err
//	//		}
//	//
//	//		if len(systems) == 0 {
//	//			return true, nil
//	//		}
//	//
//	//		Expect(len(systems)).To(Equal(1))
//	//		system := systems[0]
//	//		Expect(system.ID).To(Equal(systemID))
//	//		Expect(system.DefinitionURL).To(Equal(""))
//	//		Expect(len(system.Services)).To(Equal(0))
//	//		Expect(system.State).To(Equal(types.SystemStateDeleting))
//	//		return false, nil
//	//	})
//	//	Expect(err).NotTo(HaveOccurred())
//	//})
//})
