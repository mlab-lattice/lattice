package test

import (
	"github.com/mlab-lattice/system/pkg/definition/block"
	jsonutil "github.com/mlab-lattice/system/pkg/util/json"
)

func MockResources() *block.Resources {
	instanceType := "mock.instance.type"
	var one int32 = 1
	return &block.Resources{
		MinInstances: &one,
		MaxInstances: &one,
		InstanceType: &instanceType,
	}
}

func MockResourcesExpectedJson() []byte {
	return GenerateResourcesExpectedJson(
		[]byte(`1`),
		[]byte(`1`),
		[]byte(`"mock.instance.type"`),
	)
}

func GenerateResourcesExpectedJson(minInstances, maxInstances, instanceType []byte) []byte {
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
