package resolver

import (
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/template"
	"github.com/mlab-lattice/lattice/pkg/util/git"
)

type TemplateStore interface {
	Ready() bool
	Put(systemID v1.SystemID, ref *git.FileReference, t *template.Template) error
	Get(systemID v1.SystemID, ref *git.FileReference) (*template.Template, error)
}

type TemplateDoesNotExistError struct{}

func (e *TemplateDoesNotExistError) Error() string {
	return "template does not exist"
}

func NewMemoryTemplateStore() TemplateStore {
	return &MemoryTemplateStore{
		store: make(map[string]*template.Template),
	}
}

// MemoryTemplateStore implements a basic TemplateStore that holds the templates in memory.
type MemoryTemplateStore struct {
	store map[string]*template.Template
}

func (s *MemoryTemplateStore) Ready() bool {
	return true
}

func (s *MemoryTemplateStore) Put(systemID v1.SystemID, ref *git.FileReference, t *template.Template) error {
	s.store[s.refString(systemID, ref)] = t
	return nil
}

func (s *MemoryTemplateStore) Get(systemID v1.SystemID, ref *git.FileReference) (*template.Template, error) {
	t, ok := s.store[s.refString(systemID, ref)]
	if !ok {
		return nil, &TemplateDoesNotExistError{}
	}

	return t, nil
}

func (s *MemoryTemplateStore) refString(systemID v1.SystemID, ref *git.FileReference) string {
	return fmt.Sprintf("%v.%v.%v.%v", systemID, ref.RepositoryURL, ref.Commit, ref.File)
}
