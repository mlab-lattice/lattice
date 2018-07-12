package latticeutil

import (
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	"github.com/mlab-lattice/lattice/pkg/util/sha1"
)

func HashNodePath(path tree.NodePath) (string, error) {
	return sha1.EncodeToHexString([]byte(path.String()))
}
