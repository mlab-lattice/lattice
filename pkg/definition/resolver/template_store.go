package resolver

import (
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/definition/template"
	definitionv1 "github.com/mlab-lattice/lattice/pkg/definition/v1"
)

type TemplateStore interface {
	Put(ref *definitionv1.GitRepositoryReference, t *template.Template) error
	Get(ref *definitionv1.GitRepositoryReference) (*template.Template, error)
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

func (s *MemoryTemplateStore) Put(ref *definitionv1.GitRepositoryReference, t *template.Template) error {
	s.store[s.refString(ref)] = t
	return nil
}

func (s *MemoryTemplateStore) Get(ref *definitionv1.GitRepositoryReference) (*template.Template, error) {
	t, ok := s.store[s.refString(ref)]
	if !ok {
		return nil, &TemplateDoesNotExistError{}
	}

	return t, nil
}

func (s *MemoryTemplateStore) refString(ref *definitionv1.GitRepositoryReference) string {
	return fmt.Sprintf("%v.%v.%v", ref.URL, *ref.Commit, ref.File)
}
