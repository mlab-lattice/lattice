package system

import (
	"fmt"
	"time"

	v1client "github.com/mlab-lattice/system/pkg/api/client/v1"
	"github.com/mlab-lattice/system/pkg/api/v1"

	. "github.com/onsi/gomega"

	set "github.com/deckarep/golang-set"
)

func Create(client v1client.SystemClient, id v1.SystemID, definitionURL string) v1.SystemID {
	system, err := client.Create(id, definitionURL)
	Expect(err).NotTo(HaveOccurred(), "error creating system")

	Expect(system).To(Not(BeNil()), "returned system is nil")
	Expect(system.DefinitionURL).To(Equal(definitionURL), fmt.Sprintf("returned system %v has unexpected definition url %v (expected %v)", system.ID, system.DefinitionURL, definitionURL))
	Expect(system.State).To(Equal(v1.SystemStatePending), fmt.Sprintf("returned system %v has unexpected state %v (expected %v)"), system.ID, system.State, v1.SystemStatePending)

	return system.ID
}

func CreateSuccessfully(client v1client.SystemClient, id v1.SystemID, definitionURL string) v1.SystemID {
	systemID := Create(client, id, definitionURL)
	WaitUntilStable(client, id, 1*time.Second, 10*time.Second)
	return systemID
}

func List(client v1client.SystemClient, expectedSystems []v1.SystemID) {
	systems, err := client.List()
	Expect(err).NotTo(HaveOccurred(), "error listing systems")

	Expect(len(systems)).To(Equal(len(expectedSystems)), "system list does not have the expected number of results")

	expectedSystemsSet := set.NewSet()
	seenSystems := set.NewSet()

	for _, systemID := range expectedSystems {
		expectedSystemsSet.Add(systemID)
	}

	for _, system := range systems {
		Expect(expectedSystemsSet.Contains(system.ID)).To(BeTrue(), fmt.Sprintf("system %v is in the list but not in the list of expected systems", system.ID))

		Expect(seenSystems.Contains(system.ID)).To(BeFalse(), fmt.Sprintf("system %v was repeated in the list", system.ID))
		seenSystems.Add(system.ID)
	}
}

func Get(client v1client.SystemClient, id v1.SystemID) *v1.System {
	system, err := client.Get(id)
	Expect(err).NotTo(HaveOccurred(), "error getting system")

	Expect(system).To(Not(BeNil()), "system was nil")
	Expect(system.ID).To(Equal(id), "system had unexpected id %v (expected %v)", system.ID, id)
	return system
}

func Delete(client v1client.SystemClient, id v1.SystemID) {
	err := client.Delete(id)
	Expect(err).NotTo(HaveOccurred(), "error deleting system")
}

func DeleteSuccessfully(client v1client.SystemClient, id v1.SystemID, interval, timeout time.Duration) {
	Delete(client, id)
	WaitUntilDeleted(client, id, interval, timeout)
}
