package resolver

import (
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/resolver/template"
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
