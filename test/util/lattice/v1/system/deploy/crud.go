package deploy

import (
	"fmt"
	"time"

	v1client "github.com/mlab-lattice/lattice/pkg/api/client/v1"
	"github.com/mlab-lattice/lattice/pkg/api/v1"

	. "github.com/onsi/gomega"

	set "github.com/deckarep/golang-set"
)

func CreateFromBuild(client v1client.DeployClient, buildID v1.BuildID) v1.DeployID {
	deploy, err := client.CreateFromBuild(buildID)
	Expect(err).NotTo(HaveOccurred(), "error deploying system")

	Expect(deploy).To(Not(Equal(nil)), "returned deploy is nil")
	Expect(deploy.State).To(Equal(v1.DeployStatePending), fmt.Sprintf("deploy %v has unexpected state %v (expected %v)", deploy.ID, deploy.State, v1.DeployStatePending))

	return deploy.ID
}

func DeploySuccessfullyFromBuild(client v1client.DeployClient, version v1.SystemVersion, interval, timeout time.Duration) v1.DeployID {
	deployID := CreateFromVersion(client, version)
	WaitUntilSucceeded(client, deployID, interval, timeout)
	return deployID
}

func CreateFromVersion(client v1client.DeployClient, version v1.SystemVersion) v1.DeployID {
	deploy, err := client.CreateFromVersion(version)
	Expect(err).NotTo(HaveOccurred(), "error deploying system")

	Expect(deploy).To(Not(Equal(nil)), "returned deploy is nil")
	Expect(deploy.State).To(Equal(v1.DeployStatePending), fmt.Sprintf("deploy %v has unexpected state %v (expected %v)", deploy.ID, deploy.State, v1.DeployStatePending))

	return deploy.ID
}

func DeploySuccessfullyFromVersion(client v1client.DeployClient, version v1.SystemVersion, interval, timeout time.Duration) v1.DeployID {
	deployID := CreateFromVersion(client, version)
	WaitUntilSucceeded(client, deployID, interval, timeout)
	return deployID
}

func List(client v1client.DeployClient, expectedDeploys []v1.DeployID) {
	deploys, err := client.List()
	Expect(err).NotTo(HaveOccurred(), "error listing deploys")

	Expect(len(deploys)).To(Equal(len(expectedDeploys)), "deploy list does not have the expected number of results")

	expectedDeploySet := set.NewSet()
	seenDeploys := set.NewSet()

	for _, deployID := range expectedDeploys {
		expectedDeploySet.Add(deployID)
	}

	for _, deploy := range deploys {
		Expect(expectedDeploySet.Contains(deploy.ID)).To(BeTrue(), fmt.Sprintf("deploy %v is in the list but not in the list of expected deploys", deploy.ID))

		Expect(seenDeploys.Contains(deploy.ID)).To(BeFalse(), fmt.Sprintf("deploy %v was repeated in the list", deploy.ID))
		seenDeploys.Add(deploy.ID)
	}
}

func Get(client v1client.DeployClient, id v1.DeployID) *v1.Deploy {
	deploy, err := client.Get(id)
	Expect(err).NotTo(HaveOccurred(), "error getting deploy")

	Expect(deploy).To(Not(BeNil()), "deploy was nil")
	Expect(deploy.ID).To(Equal(id), "deploy had unexpected id %v (expected %v)", deploy.ID, id)
	return deploy
}
