package backend

import (
	"fmt"

	kubeutil "github.com/mlab-lattice/system/pkg/backend/kubernetes/util/kubernetes"
	"github.com/mlab-lattice/system/pkg/definition/tree"
	"github.com/mlab-lattice/system/pkg/types"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (kb *KubernetesBackend) ListSecrets(systemID types.SystemID) ([]types.Secret, error) {
	namespace := kubeutil.SystemNamespace(kb.clusterID, systemID)

	secrets, err := kb.kubeClient.CoreV1().Secrets(namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var externalSecrets []types.Secret
	for _, secret := range secrets.Items {
		path, err := tree.NodePathFromDomain(secret.Name)
		if err != nil {
			fmt.Printf("unexpected secret name format: %v\n", secret.Name)
			continue
		}

		for name, value := range secret.StringData {
			externalSecrets = append(externalSecrets, types.Secret{
				Path:  path,
				Name:  name,
				Value: value,
			})
		}
	}

	return externalSecrets, nil
}

func (kb *KubernetesBackend) GetSecret(id types.SystemID, path tree.NodePath, name string) (*types.Secret, bool, error) {
	namespace := kubeutil.SystemNamespace(kb.clusterID, id)
	secret, err := kb.kubeClient.CoreV1().Secrets(namespace).Get(path.ToDomain(true), metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, false, nil
		}
		return nil, false, err
	}

	value, ok := secret.StringData[name]
	if !ok {
		return nil, false, nil
	}

	externalSecret := &types.Secret{
		Path:  path,
		Name:  name,
		Value: value,
	}
	return externalSecret, true, nil
}

func (kb *KubernetesBackend) SetSecret(id types.SystemID, path tree.NodePath, name, value string) error {
	namespace := kubeutil.SystemNamespace(kb.clusterID, id)
	secret, err := kb.kubeClient.CoreV1().Secrets(namespace).Get(path.ToDomain(true), metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return kb.createSecret(id, path, name, value)
		}
		return err
	}

	secret.StringData[name] = value
	_, err = kb.kubeClient.CoreV1().Secrets(namespace).Update(secret)
	return err
}

func (kb *KubernetesBackend) createSecret(id types.SystemID, path tree.NodePath, name, value string) error {
	namespace := kubeutil.SystemNamespace(kb.clusterID, id)
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

func (kb *KubernetesBackend) UnsetSecret(id types.SystemID, path tree.NodePath, name string) error {
	namespace := kubeutil.SystemNamespace(kb.clusterID, id)
	secret, err := kb.kubeClient.CoreV1().Secrets(namespace).Get(path.ToDomain(true), metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	}

	delete(secret.StringData, name)
	if len(secret.StringData) == 0 {
		return kb.kubeClient.CoreV1().Secrets(namespace).Delete(secret.Name, nil)
	}

	_, err = kb.kubeClient.CoreV1().Secrets(namespace).Update(secret)
	return err
}
