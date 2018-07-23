package resolver

import (
	"github.com/mlab-lattice/lattice/pkg/definition/template"
	"github.com/mlab-lattice/lattice/pkg/definition/v1"
)

type TemplateStore interface {
	Put(ref *v1.GitRepositoryReference, template *template.Template) error
	Get(ref *v1.GitRepositoryReference) (*template.Template, error)
}
