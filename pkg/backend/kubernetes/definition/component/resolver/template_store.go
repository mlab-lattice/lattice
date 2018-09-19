package resolver

import (
	"encoding/json"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	latticev1 "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/apis/lattice/v1"
	latticeclientset "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/generated/clientset/versioned"
	latticeinformers "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/generated/informers/externalversions"
	latticelisters "github.com/mlab-lattice/lattice/pkg/backend/kubernetes/customresource/generated/listers/lattice/v1"
	"github.com/mlab-lattice/lattice/pkg/backend/kubernetes/util/kubernetes"
	"github.com/mlab-lattice/lattice/pkg/definition/component/resolver"
	"github.com/mlab-lattice/lattice/pkg/definition/component/resolver/template"
	"github.com/mlab-lattice/lattice/pkg/util/git"
	"github.com/mlab-lattice/lattice/pkg/util/sha1"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"

	"github.com/satori/go.uuid"
	"k8s.io/client-go/tools/cache"
)

func NewKubernetesTemplateStore(
	namespacePrefix string,
	latticeClient latticeclientset.Interface,
	latticeInformerFactory latticeinformers.SharedInformerFactory,
	stopCh <-chan struct{},
) *KubernetesTemplateStore {
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
}

func (s *KubernetesTemplateStore) Ready() bool {
	return cache.WaitForCacheSync(s.stopCh, s.gitTemplateListerHasSynced, s.templateListerHasSynced)
}

func (s *KubernetesTemplateStore) Put(
	systemID v1.SystemID,
	ref *git.FileReference,
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

	lt, err := s.putTemplate(systemID, t)
	if err != nil {
		return err
	}

	_, err = s.putGitTemplate(systemID, ref, lt)
	return err
}

func (s *KubernetesTemplateStore) Get(systemID v1.SystemID, ref *git.FileReference) (*template.Template, error) {
	lgt, err := s.gitTemplateFromLister(systemID, ref)
	if err != nil {
		return nil, err
	}

	namespace := kubernetes.SystemNamespace(s.namespacePrefix, systemID)
	lt, err := s.templateLister.Templates(namespace).Get(lgt.Spec.TemplateDigest)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, &resolver.TemplateDoesNotExistError{}
		}

		return nil, err
	}

	return lt.Spec.Template, nil
}

func (s *KubernetesTemplateStore) gitTemplateFromLister(systemID v1.SystemID, ref *git.FileReference) (*latticev1.GitTemplate, error) {
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

	// It's possible that a race lost and we created two different git templates at the same time.
	// In theory since they should contain the same contents, they should be identical.
	// It seems easiest to handle the small chance of this race happening by simply ignoring it
	// and creating an extra small resource rather than trying to prevent it.
	// An alternate technique could have been to hash the ref's contents, but then we can't change
	// the git.FileReference struct at all without having our hashes be invalidated.
	return gitTemplates[0], nil
}

func (s *KubernetesTemplateStore) gitTemplateSelector(ref *git.FileReference) (labels.Selector, error) {
	l, err := s.gitTemplateLabels(ref)
	if err != nil {
		return nil, err
	}

	selector := labels.NewSelector()
	for k, v := range l {
		requirement, err := labels.NewRequirement(k, selection.Equals, []string{v})
		if err != nil {
			return nil, err
		}

		selector = selector.Add(*requirement)
	}

	return selector, nil
}

func (s *KubernetesTemplateStore) putTemplate(systemID v1.SystemID, t *template.Template) (*latticev1.Template, error) {
	// If there wasn't a git template already, first get the digest of the template
	data, err := json.Marshal(&t)
	if err != nil {
		return nil, err
	}

	digest, err := sha1.EncodeToHexString(data)
	if err != nil {
		return nil, err
	}

	// Check to see if the template already exists in our cache
	var lt *latticev1.Template
	namespace := kubernetes.SystemNamespace(s.namespacePrefix, systemID)
	lt, err = s.templateLister.Templates(namespace).Get(digest)
	if err != nil {
		// If there was an error other than the template not being in the cache,
		// return it.
		if !errors.IsNotFound(err) {
			return nil, err
		}

		// Try to create the template
		lt = &latticev1.Template{
			ObjectMeta: metav1.ObjectMeta{
				Name: digest,
			},
			Spec: latticev1.TemplateSpec{
				Template: t,
			},
		}

		// If we get an AlreadyExists error, we lost a race, which is fine
		// since it means the template exists now
		lt, err = s.latticeClient.LatticeV1().Templates(namespace).Create(lt)
		if err != nil && !errors.IsAlreadyExists(err) {
			return nil, err
		}
	}

	return lt, nil
}

func (s *KubernetesTemplateStore) putGitTemplate(
	systemID v1.SystemID,
	ref *git.FileReference,
	lt *latticev1.Template,
) (*latticev1.GitTemplate, error) {
	l, err := s.gitTemplateLabels(ref)
	if err != nil {
		return nil, err
	}

	lgt := &latticev1.GitTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Name:   uuid.NewV4().String(),
			Labels: l,
		},
		Spec: latticev1.GitTemplateSpec{
			TemplateDigest: lt.Name,
		},
	}

	namespace := kubernetes.SystemNamespace(s.namespacePrefix, systemID)
	return s.latticeClient.LatticeV1().GitTemplates(namespace).Create(lgt)
}

func (s *KubernetesTemplateStore) gitTemplateLabels(ref *git.FileReference) (map[string]string, error) {
	urlHash, err := sha1.EncodeToHexString([]byte(ref.RepositoryURL))
	if err != nil {
		return nil, err
	}

	fileHash, err := sha1.EncodeToHexString([]byte(ref.File))
	if err != nil {
		return nil, err
	}

	m := map[string]string{
		latticev1.GitTemplateRepoURLLabelKey: urlHash,
		latticev1.GitTemplateCommitLabelKey:  ref.Commit,
		latticev1.GitTemplateFileLabelKey:    fileHash,
	}
	return m, nil
}
