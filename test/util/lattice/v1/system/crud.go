package system

import (
	"fmt"
	"time"

	v1client "github.com/mlab-lattice/system/pkg/api/client/v1"
	"github.com/mlab-lattice/system/pkg/api/v1"
	"github.com/mlab-lattice/system/test/util/lattice/v1/system/expected"

	. "github.com/onsi/gomega"
	gomegatypes "github.com/onsi/gomega/types"

	set "github.com/deckarep/golang-set"
)

func Create(client v1client.SystemClient, id v1.SystemID, definitionURL string) v1.SystemID {
	system, err := client.Create(id, definitionURL)
	Expect(err).NotTo(HaveOccurred(), "error creating system")

	Expect(system).To(Not(Equal(nil)), "returned system is nil")

	expectedSystem := &expected.System{
		ID:            id,
		DefinitionURL: definitionURL,
		ValidServices: nil,
		ValidStates:   []v1.SystemState{v1.SystemStatePending},
	}

	CompareSystems(system, expectedSystem)

	return system.ID
}

func CreateSuccesfully(client v1client.SystemClient, id v1.SystemID, definitionURL string) v1.SystemID {
	systemID := Create(client, id, definitionURL)
	WaitUntilStable(client, id, 1*time.Second, 10*time.Second)
	return systemID
}

func List(client v1client.SystemClient, expectedSystems []expected.System) {
	systems, err := client.List()
	Expect(err).NotTo(HaveOccurred(), "error listing systems")

	Expect(len(systems)).To(Equal(len(expectedSystems)), "system list does not have the expected number of results")

	expectedSystemsMap := make(map[v1.SystemID]expected.System)
	seenSystems := set.NewSet()

	for _, system := range expectedSystems {
		expectedSystemsMap[system.ID] = system
	}

	for _, system := range systems {
		expectedSystem, ok := expectedSystemsMap[system.ID]
		Expect(ok).To(BeTrue(), fmt.Sprintf("system %v is in the list but not in the list of expected systems", system.ID))

		Expect(seenSystems.Contains(system.ID)).To(BeFalse(), fmt.Sprintf("system %v was repeated in the list", system.ID))
		seenSystems.Add(system.ID)

		CompareSystems(&system, &expectedSystem)
	}
}

func Get(client v1client.SystemClient, id v1.SystemID, expectedSystem *expected.System) *v1.System {
	system, err := client.Get(id)
	Expect(err).NotTo(HaveOccurred(), "error getting system")

	CompareSystems(system, expectedSystem)
	return system
}

func Delete(client v1client.SystemClient, id v1.SystemID) {
	err := client.Delete(id)
	Expect(err).NotTo(HaveOccurred(), "error deleting system")
}

func DeleteSuccesfully(client v1client.SystemClient, id v1.SystemID, interval, timeout time.Duration) {
	Delete(client, id)
	WaitUntilDeleted(client, id, 1*time.Second, 10*time.Second)
}

func CompareSystems(system *v1.System, expected *expected.System) {
	Expect(system.ID).To(Equal(expected.ID), fmt.Sprintf("system has unexpected ID %v (expected %v)", system.ID, expected.ID))
	Expect(system.DefinitionURL).To(Equal(expected.DefinitionURL), fmt.Sprintf("system %v has unexpected definition url %v (expected %v)", system.ID, system.DefinitionURL, expected.DefinitionURL))
	Expect(len(system.Services)).To(Equal(len(expected.ValidServices)), fmt.Sprintf("system %v has unexpected number of systems %v (expected %v)", system.ID, len(system.Services), len(expected.DesiredServices)))

	var validStates []gomegatypes.GomegaMatcher
	for _, validState := range expected.ValidStates {
		validStates = append(validStates, Equal(validState))
	}
	Expect(system.State).To(SatisfyAny(validStates...), fmt.Sprintf("system %v has unexpected state %v (expected one of %v)", system.ID, system.State, expected.ValidStates))
}
