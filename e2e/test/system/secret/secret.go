package build

import (
	"time"

	"github.com/mlab-lattice/lattice/e2e/test/context"
	. "github.com/mlab-lattice/lattice/e2e/util/ginkgo"
	"github.com/mlab-lattice/lattice/e2e/util/lattice/v1/system"
	"github.com/mlab-lattice/lattice/e2e/util/lattice/v1/system/secret"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("secret", func() {
	systemName := v1.SystemID("e2e-system-secret-1")
	systemURL := "https://github.com/mlab-lattice/testing__system.git"

	var systemID v1.SystemID
	It("should be able to create a system", func() {
		systemID = system.CreateSuccessfully(context.TestContext.LatticeAPIClient.V1().Systems(), systemName, systemURL)
	})

	ifSystemCreated := If("system creation succeeded", func() bool { return systemID != "" })

	ConditionallyIt(
		"should be able to list secrets, but the list should be empty",
		ifSystemCreated,
		func() {
			secret.List(context.TestContext.LatticeAPIClient.V1().Systems().Secrets(systemID), nil)
		},
	)

	secretPath := tree.Path("/a/b")
	secretName := "buzz"
	secretValue := "foobar"
	setSecret := false
	ConditionallyIt(
		"should be able to set a secret",
		ifSystemCreated,
		func() {
			secret.Set(context.TestContext.LatticeAPIClient.V1().Systems().Secrets(systemID), secretPath, secretName, secretValue)
			setSecret = true
		},
	)

	ifSecretSet := If("secret was set", func() bool { return setSecret })
	ConditionallyIt(
		"should be able to get the set secret",
		ifSecretSet,
		func() {
			value := secret.Get(context.TestContext.LatticeAPIClient.V1().Systems().Secrets(systemID), secretPath, secretName)
			Expect(value).To(Equal(secretValue))
		},
	)

	ConditionallyIt(
		"should be able to list secrets, and the list should contain the set secret",
		ifSecretSet,
		func() {
			secret.List(context.TestContext.LatticeAPIClient.V1().Systems().Secrets(systemID), []v1.Secret{
				{
					Path:  secretPath,
					Name:  secretName,
					Value: secretValue,
				},
			})
		},
	)

	ConditionallyIt(
		"should be able to unset the set secret",
		ifSecretSet,
		func() {
			secret.Unset(context.TestContext.LatticeAPIClient.V1().Systems().Secrets(systemID), secretPath, secretName)
		},
	)

	ConditionallyIt(
		"should be able to list secrets, but the list should be empty",
		ifSecretSet,
		func() {
			secret.List(context.TestContext.LatticeAPIClient.V1().Systems().Secrets(systemID), nil)
		},
	)

	ConditionallyIt(
		"should be able to delete the system",
		ifSystemCreated,
		func() {
			system.DeleteSuccessfully(context.TestContext.LatticeAPIClient.V1().Systems(), systemID, 1*time.Second, 10*time.Second)
		},
	)
})
