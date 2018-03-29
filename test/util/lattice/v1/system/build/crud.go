package build

import (
	"fmt"
	"time"

	v1client "github.com/mlab-lattice/system/pkg/api/client/v1"
	"github.com/mlab-lattice/system/pkg/api/v1"

	. "github.com/onsi/gomega"

	set "github.com/deckarep/golang-set"
)

func Create(client v1client.BuildClient, version v1.SystemVersion) v1.BuildID {
	build, err := client.Create(version)
	Expect(err).NotTo(HaveOccurred(), "error building system")

	Expect(build).To(Not(Equal(nil)), "returned build is nil")
	Expect(build.Version).To(Equal(version), fmt.Sprintf("build %v has unexpected version %v (expected %v)", build.ID, build.Version, version))
	Expect(build.State).To(Equal(v1.BuildStatePending), fmt.Sprintf("build %v has unexpected state %v (expected %v)", build.ID, build.State, v1.BuildStatePending))

	return build.ID
}

func BuildSuccessfully(client v1client.BuildClient, version v1.SystemVersion, interval, timeout time.Duration) v1.BuildID {
	buildID := Create(client, version)
	WaitUntilSucceeded(client, buildID, interval, timeout)
	return buildID
}

func List(client v1client.BuildClient, expectedBuilds []v1.BuildID) {
	builds, err := client.List()
	Expect(err).NotTo(HaveOccurred(), "error listing builds")

	Expect(len(builds)).To(Equal(len(expectedBuilds)), "build list does not have the expected number of results")

	expectedBuildsSet := set.NewSet()
	seenBuilds := set.NewSet()

	for _, buildID := range expectedBuilds {
		expectedBuildsSet.Add(buildID)
	}

	for _, build := range builds {
		Expect(expectedBuildsSet.Contains(build.ID)).To(BeTrue(), fmt.Sprintf("build %v is in the list but not in the list of expected builds", build.ID))

		Expect(seenBuilds.Contains(build.ID)).To(BeFalse(), fmt.Sprintf("build %v was repeated in the list", build.ID))
		seenBuilds.Add(build.ID)
	}
}

func Get(client v1client.BuildClient, id v1.BuildID) *v1.Build {
	build, err := client.Get(id)
	Expect(err).NotTo(HaveOccurred(), "error getting build")

	Expect(build).To(Not(BeNil()), "build was nil")
	Expect(build.ID).To(Equal(id), "build had unexpected id %v (expected %v)", build.ID, id)
	return build
}
