package backend

import (
	"fmt"

	"github.com/mlab-lattice/system/pkg/api/v1"
	kubeutil "github.com/mlab-lattice/system/pkg/backend/kubernetes/util/kubernetes"
	"github.com/mlab-lattice/system/pkg/definition/tree"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (kb *KubernetesBackend) ListSecrets(systemID v1.SystemID) ([]v1.Secret, error) {
	namespace := kubeutil.SystemNamespace(kb.latticeID, systemID)

	secrets, err := kb.kubeClient.CoreV1().Secrets(namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var externalSecrets []v1.Secret
	for _, secret := range secrets.Items {
		path, err := tree.NodePathFromDomain(secret.Name)
		if err != nil {
			fmt.Printf("unexpected secret name format: %v\n", secret.Name)
			continue
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

func (kb *KubernetesBackend) GetSecret(id v1.SystemID, path tree.NodePath, name string) (*v1.Secret, bool, error) {
	namespace := kubeutil.SystemNamespace(kb.latticeID, id)
	secret, err := kb.kubeClient.CoreV1().Secrets(namespace).Get(path.ToDomain(true), metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, false, nil
		}
		return nil, false, err
	}

	value, ok := secret.Data[name]
	if !ok {
		return nil, false, nil
	}

	externalSecret := &v1.Secret{
		Path:  path,
		Name:  name,
		Value: string(value),
	}
	return externalSecret, true, nil
}

func (kb *KubernetesBackend) SetSecret(id v1.SystemID, path tree.NodePath, name, value string) error {
	namespace := kubeutil.SystemNamespace(kb.latticeID, id)
	secret, err := kb.kubeClient.CoreV1().Secrets(namespace).Get(path.ToDomain(true), metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return kb.createSecret(id, path, name, value)
		}
		return err
	}

	secret.StringData = map[string]string{
		name: value,
	}
	_, err = kb.kubeClient.CoreV1().Secrets(namespace).Update(secret)
	return err
}

func (kb *KubernetesBackend) createSecret(id v1.SystemID, path tree.NodePath, name, value string) error {
	namespace := kubeutil.SystemNamespace(kb.latticeID, id)
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: path.ToDomain(true),
		},
		StringData: map[string]string{
			name: value,
		},
	}
	_, err := kb.kubeClient.CoreV1().Secrets(namespace).Create(secret)
	return err
}

func (kb *KubernetesBackend) UnsetSecret(id v1.SystemID, path tree.NodePath, name string) error {
	namespace := kubeutil.SystemNamespace(kb.latticeID, id)
	secret, err := kb.kubeClient.CoreV1().Secrets(namespace).Get(path.ToDomain(true), metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	}

	delete(secret.Data, name)
	if len(secret.Data) == 0 {
		return kb.kubeClient.CoreV1().Secrets(namespace).Delete(secret.Name, nil)
	}

	_, err = kb.kubeClient.CoreV1().Secrets(namespace).Update(secret)
	return err
}
