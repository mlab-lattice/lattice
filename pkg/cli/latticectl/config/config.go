package config

import (
	"github.com/mlab-lattice/system/pkg/types"
)

type Config struct {
	Context *Context `json:"context"`
}

type Context struct {
	Lattice string         `json:"lattice"`
	System  types.SystemID `json:"system"`
}
