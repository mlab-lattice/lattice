package backend

import (
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	kubeutil "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/kubernetes"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"

	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/constants"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
)

func (kb *KubernetesBackend) ListSystemSecrets(systemID v1.SystemID) ([]v1.Secret, error) {
	// ensure the system exists
	if err := kb.ensureSystemCreated(systemID); err != nil {
		return nil, err
	}

	namespace := kubeutil.SystemNamespace(kb.latticeID, systemID)

	// There are secrets in the namespace that are not secrets set for lattice.
	// Don't expose those in ListSystemSecrets
	selector := labels.NewSelector()
	requirement, err := labels.NewRequirement(constants.LabelKeySecret, selection.Exists, nil)
	if err != nil {
		return nil, err
	}
	selector = selector.Add(*requirement)

	secrets, err := kb.kubeClient.CoreV1().Secrets(namespace).List(metav1.ListOptions{LabelSelector: selector.String()})
	if err != nil {
		return nil, err
	}

	externalSecrets := make([]v1.Secret, 0)
	for _, secret := range secrets.Items {
		path, err := tree.NodePathFromDomain(secret.Name)
		if err != nil {
			return nil, err
		}

		for name, value := range secret.Data {
			externalSecrets = append(externalSecrets, v1.Secret{
				Path:  path,
				Name:  name,
				Value: string(value),
			})
		}
	}

	return externalSecrets, nil
}

func (kb *KubernetesBackend) GetSystemSecret(systemID v1.SystemID, path tree.NodePath, name string) (*v1.Secret, error) {
	// ensure the system exists
	if err := kb.ensureSystemCreated(systemID); err != nil {
		return nil, err
	}

	namespace := kubeutil.SystemNamespace(kb.latticeID, systemID)
	secret, err := kb.kubeClient.CoreV1().Secrets(namespace).Get(path.ToDomain(), metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, v1.NewInvalidSystemSecretError(path, name)
		}

		return nil, err
	}

	value, ok := secret.Data[name]
	if !ok {
		return nil, v1.NewInvalidSystemSecretError(path, name)
	}

	externalSecret := &v1.Secret{
		Path:  path,
		Name:  name,
		Value: string(value),
	}
	return externalSecret, nil
}

func (kb *KubernetesBackend) SetSystemSecret(systemID v1.SystemID, path tree.NodePath, name, value string) error {
	// ensure the system exists
	if err := kb.ensureSystemCreated(systemID); err != nil {
		return err
	}

	namespace := kubeutil.SystemNamespace(kb.latticeID, systemID)
	secret, err := kb.kubeClient.CoreV1().Secrets(namespace).Get(path.ToDomain(), metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return kb.createSecret(systemID, path, name, value)
		}

		return err
	}

	secret.StringData = map[string]string{
		name: value,
	}
	_, err = kb.kubeClient.CoreV1().Secrets(namespace).Update(secret)
	if err == nil {
		return nil
	}

	// if there was a conflict or the secret no longer exists (i.e. it was deleted since we found it)
	// return a conflict error
	if errors.IsConflict(err) || errors.IsNotFound(err) {
		return v1.NewConflictError("")
	}

	return err
}

func (kb *KubernetesBackend) createSecret(id v1.SystemID, path tree.NodePath, name, value string) error {
	namespace := kubeutil.SystemNamespace(kb.latticeID, id)
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: path.ToDomain(),
			Labels: map[string]string{
				constants.LabelKeySecret: "true",
			},
		},
		StringData: map[string]string{
			name: value,
		},
	}

	_, err := kb.kubeClient.CoreV1().Secrets(namespace).Create(secret)
	return err
}

func (kb *KubernetesBackend) UnsetSystemSecret(systemID v1.SystemID, path tree.NodePath, name string) error {
	// ensure the system exists
	if err := kb.ensureSystemCreated(systemID); err != nil {
		return err
	}

	namespace := kubeutil.SystemNamespace(kb.latticeID, systemID)
	secret, err := kb.kubeClient.CoreV1().Secrets(namespace).Get(path.ToDomain(), metav1.GetOptions{})
	if err != nil {
		// If we can't find the secret, then it is unset
		if errors.IsNotFound(err) {
			return nil
		}

		return err
	}

	delete(secret.Data, name)
	if len(secret.Data) == 0 {
		err = kb.kubeClient.CoreV1().Secrets(namespace).Delete(secret.Name, nil)
		if err != nil {
			return nil
		}

		if errors.IsConflict(err) {
			return v1.NewConflictError("")
		}

		return err
	}

	_, err = kb.kubeClient.CoreV1().Secrets(namespace).Update(secret)
	if err == nil {
		return nil
	}

	if errors.IsConflict(err) {
		return v1.NewConflictError("")
	}

	return err
}
