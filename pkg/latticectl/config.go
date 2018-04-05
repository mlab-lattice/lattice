package latticectl

import (
	"github.com/mlab-lattice/lattice/pkg/api/v1"
)

type Config struct {
	Context *ConfigContext `json:"context"`
}

type ConfigContext struct {
	Lattice string      `json:"lattice"`
	System  v1.SystemID `json:"system"`
}
