package resolver

import (
	"encoding/json"
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	latticeclientset "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/generated/clientset/versioned"
	latticeinformers "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/generated/informers/externalversions"
	latticelisters "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/generated/listers/lattice/v1"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/kubernetes"
	"github.com/mlab-lattice/lattice/pkg/definition/resolver"
	"github.com/mlab-lattice/lattice/pkg/definition/template"
	definitionv1 "github.com/mlab-lattice/lattice/pkg/definition/v1"
	"github.com/mlab-lattice/lattice/pkg/util/sha1"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/mlab-lattice/lattice/pkg/util/git"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/client-go/tools/cache"
	"sync"
)

func NewKubernetesTemplateStore(
	namespacePrefix string,
	latticeClient latticeclientset.Interface,
	latticeInformerFactory latticeinformers.SharedInformerFactory,
	stopCh <-chan struct{},
) resolver.TemplateStore {
	s := &KubernetesTemplateStore{
		namespacePrefix: namespacePrefix,
		latticeClient:   latticeClient,
		stopCh:          stopCh,

		gitTemplateLister:          latticeInformerFactory.Lattice().V1().GitTemplates().Lister(),
		gitTemplateListerHasSynced: latticeInformerFactory.Lattice().V1().GitTemplates().Informer().HasSynced,

		templateLister:          latticeInformerFactory.Lattice().V1().Templates().Lister(),
		templateListerHasSynced: latticeInformerFactory.Lattice().V1().Templates().Informer().HasSynced,
	}

	latticeInformerFactory.Start(stopCh)
	return s
}

// MemoryTemplateStore implements a TemplateStore that uses custom resources to store templates.
type KubernetesTemplateStore struct {
	namespacePrefix string
	latticeClient   latticeclientset.Interface
	stopCh          <-chan struct{}

	gitTemplateLister          latticelisters.GitTemplateLister
	gitTemplateListerHasSynced cache.InformerSynced

	templateLister          latticelisters.TemplateLister
	templateListerHasSynced cache.InformerSynced

	insertLock sync.Mutex
}

func (s *KubernetesTemplateStore) Ready() bool {
	return cache.WaitForCacheSync(s.stopCh, s.templateListerHasSynced)
}

func (s *KubernetesTemplateStore) Put(
	systemID v1.SystemID,
	ref git.FileReference,
	t *template.Template,
) error {
	// If the git template already exists, return with no error
	_, err := s.gitTemplateFromLister(systemID, ref)
	if err == nil {
		return nil
	}

	// If the error is not a resolver.TemplateDoesNotExistError, return the error
	if _, ok := err.(*resolver.TemplateDoesNotExistError); !ok {
		return err
	}

	// If there wasn't a git template already, first get the digest of the template
	data, err := json.Marshal(&t)
	if err != nil {
		return err
	}

	digest, err := sha1.EncodeToHexString(data)
	if err != nil {
		return err
	}

	// Check to see if the template already exists in our cache
	namespace := kubernetes.SystemNamespace(s.namespacePrefix, systemID)
	_, err = s.templateLister.Templates(namespace).Get(digest)
	if err != nil {
		// If there was an error other than the template not being in the cache,
		// return it.
		if !errors.IsNotFound(err) {
			return err
		}

		// Try to create the template
		lt := &latticev1.Template{
			ObjectMeta: metav1.ObjectMeta{
				Name: digest,
			},
			Spec: t,
		}

		// If we get an AlreadyExists error, we lost a race, which is fine
		// since it means the template exists now
		_, err = s.latticeClient.LatticeV1().Templates(namespace).Create(lt)
		if err != nil && !errors.IsAlreadyExists(err) {
			return err
		}
	}

	// basic implementation outline idea:
	// 	hash the template, and look to see if a template.lattice.mlab.com with the hash already exists. if not, create it
	//  look to see if a gittemplate.lattice.mlab.com exists for the reference. if not, create it pointing to and as
	//    an owner of the template.lattice.mlab.com
	// later, when creating a build, evaluate the template and store it as a definition.lattice.mlab.com, and point
	//   the build at it?
	return nil
}

func (s *KubernetesTemplateStore) Get(systemID v1.SystemID, reference git.FileReference) (*template.Template, error) {
	return nil, &resolver.TemplateDoesNotExistError{}
}

func (s *KubernetesTemplateStore) gitTemplateFromLister(systemID v1.SystemID, ref git.FileReference) (*latticev1.GitTemplate, error) {
	selector, err := s.gitTemplateSelector(ref)
	if err != nil {
		return nil, err
	}

	namespace := kubernetes.SystemNamespace(s.namespacePrefix, systemID)
	gitTemplates, err := s.gitTemplateLister.GitTemplates(namespace).List(selector)
	if err != nil {
		return nil, err
	}

	if len(gitTemplates) == 0 {
		return nil, &resolver.TemplateDoesNotExistError{}
	}

	if len(gitTemplates) > 1 {
		return nil, fmt.Errorf("found multiple templates for repo %v commit %v file %v", ref.RepositoryURL, ref.Commit, ref.File)
	}

	return gitTemplates[0], nil
}

func (s *KubernetesTemplateStore) gitTemplateSelector(ref git.FileReference) (labels.Selector, error) {
	repoURLRequirement, err := labels.NewRequirement(latticev1.GitTemplateRepoURLLabelKey, selection.Equals, []string{ref.RepositoryURL})
	if err != nil {
		return nil, err
	}

	commitRequirement, err := labels.NewRequirement(latticev1.GitTemplateCommitLabelKey, selection.Equals, []string{ref.Commit})
	if err != nil {
		return nil, err
	}

	fileRequirement, err := labels.NewRequirement(latticev1.GitTemplateCommitLabelKey, selection.Equals, []string{ref.File})
	if err != nil {
		return nil, err
	}

	selector := labels.NewSelector()
	selector = selector.Add(*repoURLRequirement, *commitRequirement, *fileRequirement)
	return selector, nil
}
