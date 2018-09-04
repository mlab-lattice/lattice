package secret

import (
	"fmt"
	//"time"

	v1client "github.com/mlab-lattice/lattice/pkg/api/client/v1"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"

	. "github.com/onsi/gomega"

	set "github.com/deckarep/golang-set"
)

func List(client v1client.SystemSecretClient, expectedSecrets []v1.Secret) {
	secrets, err := client.List()
	Expect(err).NotTo(HaveOccurred(), "error listing secrets")

	Expect(len(secrets)).To(Equal(len(expectedSecrets)), "secret list does not have the expected number of results")

	expectedSecretsMap := make(map[tree.PathSubcomponent]string)
	seenSecrets := set.NewSet()

	for _, secret := range expectedSecrets {
		expectedSecretsMap[secret.Path] = secret.Value
	}

	for _, secret := range secrets {
		value, ok := expectedSecretsMap[secret.Path]
		Expect(ok).To(BeTrue(), fmt.Sprintf("secret %v is in the list but not in the list of expected secrets", secret.Path.String()))
		Expect(secret.Value).To(Equal(value), "secret %v did not have expected value", secret.Path.String())

		Expect(seenSecrets.Contains(secret.Path)).To(BeFalse(), fmt.Sprintf("secret %v was repeated in the list", secret.Path.String()))
		seenSecrets.Add(secret.Path)
	}
}

func Set(client v1client.SystemSecretClient, path tree.PathSubcomponent, value string) {
	err := client.Set(path, value)
	Expect(err).NotTo(HaveOccurred(), "error getting build")
}

func Get(client v1client.SystemSecretClient, path tree.PathSubcomponent) string {
	secret, err := client.Get(path)
	Expect(err).NotTo(HaveOccurred(), "error getting build")
	Expect(secret.Path).To(Equal(path))
	return secret.Value
}

func Unset(client v1client.SystemSecretClient, path tree.PathSubcomponent) {
	err := client.Unset(path)
	Expect(err).NotTo(HaveOccurred(), "error getting build")
}
