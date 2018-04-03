package secret

import (
	"fmt"
	//"time"

	v1client "github.com/mlab-lattice/system/pkg/api/client/v1"
	"github.com/mlab-lattice/system/pkg/api/v1"
	"github.com/mlab-lattice/system/pkg/definition/tree"

	. "github.com/onsi/gomega"

	set "github.com/deckarep/golang-set"
)

func List(client v1client.SecretClient, expectedSecrets []v1.Secret) {
	secrets, err := client.List()
	Expect(err).NotTo(HaveOccurred(), "error listing secrets")

	Expect(len(secrets)).To(Equal(len(expectedSecrets)), "secret list does not have the expected number of results")

	expectedSecretsMap := make(map[string]string)
	seenSecrets := set.NewSet()

	for _, secret := range expectedSecrets {
		expectedSecretsMap[fmt.Sprintf("%v:%v", secret.Path, secret.Name)] = secret.Value
	}

	for _, secret := range secrets {
		secretID := fmt.Sprintf("%v:%v", secret.Path, secret.Name)
		value, ok := expectedSecretsMap[secretID]
		Expect(ok).To(BeTrue(), fmt.Sprintf("secret %v is in the list but not in the list of expected secrets", secretID))
		Expect(secret.Value).To(Equal(value), "secret %v did not have expected value", secretID)

		Expect(seenSecrets.Contains(secretID)).To(BeFalse(), fmt.Sprintf("secret %v was repeated in the list", secretID))
		seenSecrets.Add(secretID)
	}
}

func Set(client v1client.SecretClient, path tree.NodePath, name, value string) {
	err := client.Set(path, name, value)
	Expect(err).NotTo(HaveOccurred(), "error getting build")
}

func Get(client v1client.SecretClient, path tree.NodePath, name string) string {
	secret, err := client.Get(path, name)
	Expect(err).NotTo(HaveOccurred(), "error getting build")
	Expect(secret.Path).To(Equal(path))
	Expect(secret.Name).To(Equal(name))
	return secret.Value
}

func Unset(client v1client.SecretClient, path tree.NodePath, name string) {
	err := client.Unset(path, name)
	Expect(err).NotTo(HaveOccurred(), "error getting build")
}
