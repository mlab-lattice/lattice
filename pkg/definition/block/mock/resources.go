package mock

import (
	"github.com/mlab-lattice/lattice/pkg/definition/block"
	jsonutil "github.com/mlab-lattice/lattice/pkg/util/json"
)

func Resources() *block.Resources {
	instanceType := "mock.instance.type"
	var one int32 = 1
	return &block.Resources{
		MinInstances: &one,
		MaxInstances: &one,
		InstanceType: &instanceType,
	}
}

func ResourcesExpectedJSON() []byte {
	return GenerateResourcesExpectedJSON(
		[]byte(`1`),
		[]byte(`1`),
		[]byte(`"mock.instance.type"`),
	)
}

func GenerateResourcesExpectedJSON(minInstances, maxInstances, instanceType []byte) []byte {
	return jsonutil.GenerateObjectBytes([]jsonutil.FieldBytes{
		{
			Name:  "min_instances",
			Bytes: minInstances,
		},
		{
			Name:  "max_instances",
			Bytes: maxInstances,
		},
		{
			Name:      "instance_type",
			Bytes:     instanceType,
			OmitEmpty: true,
		},
	})
}
