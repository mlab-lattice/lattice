package config

import (
	"github.com/mlab-lattice/lattice/pkg/api/v1"
)

type Config struct {
	Context *Context `json:"context"`
}

type Context struct {
	Lattice string      `json:"lattice"`
	System  v1.SystemID `json:"system"`
}
