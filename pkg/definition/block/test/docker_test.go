package test

import (
	"reflect"
	"testing"

	"github.com/mlab-lattice/system/pkg/definition/block"
)

func TestDockerImage_Validate(t *testing.T) {

}

func TestDockerImage_JSON(t *testing.T) {
	JSON(
		t,
		reflect.TypeOf(block.DockerImage{}),
		[]JSONTest{
			{
				Description: "MockDockerImage",
				Bytes:       MockDockerImageExpectedJSON(),
				ValuePtr:    MockDockerImage(),
			},
		},
	)
}
