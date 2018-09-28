package resolver

import (
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/kubernetes"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/latticeutil"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"

	"github.com/mlab-lattice/lattice/pkg/definition/resolver"
	"k8s.io/apimachinery/pkg/api/errors"
	kubeinformers "k8s.io/client-go/informers"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
)

func NewKubernetesSecretStore(
	namespacePrefix string,
	kubeInformerFactory kubeinformers.SharedInformerFactory,
	stopCh <-chan struct{},
) *KubernetesSecretStore {
	s := &KubernetesSecretStore{
		namespacePrefix: namespacePrefix,
		stopCh:          stopCh,

		secretLister:          kubeInformerFactory.Core().V1().Secrets().Lister(),
		secretListerHasSynced: kubeInformerFactory.Core().V1().Secrets().Informer().HasSynced,
	}

	kubeInformerFactory.Start(stopCh)
	return s
}

// MemorySecretStore implements a SecretStore that uses custom resources to store secrets.
type KubernetesSecretStore struct {
	namespacePrefix string
	stopCh          <-chan struct{}

	secretLister          corelisters.SecretLister
	secretListerHasSynced cache.InformerSynced
}

func (s *KubernetesSecretStore) Ready() bool {
	return cache.WaitForCacheSync(s.stopCh, s.secretListerHasSynced)
}

func (s *KubernetesSecretStore) Get(systemID v1.SystemID, path tree.PathSubcomponent) (string, error) {
	name, err := latticeutil.HashPath(path.Path())
	if err != nil {
		return "", err
	}

	namespace := kubernetes.SystemNamespace(s.namespacePrefix, systemID)
	secret, err := s.secretLister.Secrets(namespace).Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			return "", &resolver.SecretDoesNotExistError{}
		}
		return "", err
	}

	data, ok := secret.Data[path.Subcomponent()]
	if !ok {
		return "", &resolver.SecretDoesNotExistError{}
	}

	return string(data), nil
}
