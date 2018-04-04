package test

import (
	"reflect"
	"testing"

	"github.com/mlab-lattice/lattice/pkg/definition/block"
)

func TestResources_Validate(t *testing.T) {
	instanceType := "instance-type"
	var one int32 = 1
	var two int32 = 2
	Validate(
		t,
		nil,

		// Invalid Builds
		[]ValidateTest{
			{
				Description:     "empty",
				DefinitionBlock: &block.Resources{},
			},
			{
				Description: "NumInstances and MaxInstances",
				DefinitionBlock: &block.Resources{
					NumInstances: &two,
					MaxInstances: &one,
				},
			},
			{
				Description: "NumInstances and MinInstances",
				DefinitionBlock: &block.Resources{
					NumInstances: &two,
					MinInstances: &one,
				},
			},
			{
				Description: "NumInstances, MinInstances, and MaxInstances",
				DefinitionBlock: &block.Resources{
					NumInstances: &two,
					MinInstances: &one,
					MaxInstances: &one,
				},
			},
			{
				Description: "MaxInstances < MinInstances",
				DefinitionBlock: &block.Resources{
					MinInstances: &two,
					MaxInstances: &one,
					InstanceType: &instanceType,
				},
			},
		},

		// Valid Builds
		[]ValidateTest{
			{
				Description: "MinInstances == MaxInstances, no InstanceType",
				DefinitionBlock: &block.Resources{
					MinInstances: &one,
					MaxInstances: &one,
				},
			},
			{
				Description: "MinInstances < MaxInstances, no InstanceType",
				DefinitionBlock: &block.Resources{
					MinInstances: &one,
					MaxInstances: &two,
				},
			},
			{
				Description: "MinInstances == MaxInstances, with InstanceType",
				DefinitionBlock: &block.Resources{
					MinInstances: &one,
					MaxInstances: &one,
					InstanceType: &instanceType,
				},
			},
			{
				Description: "MinInstances < MaxInstances, no InstanceType",
				DefinitionBlock: &block.Resources{
					MinInstances: &one,
					MaxInstances: &two,
					InstanceType: &instanceType,
				},
			},
			{
				Description: "NumInstances, no InstanceType",
				DefinitionBlock: &block.Resources{
					NumInstances: &one,
				},
			},
			{
				Description: "MinInstances == MaxInstances, with InstanceType",
				DefinitionBlock: &block.Resources{
					NumInstances: &one,
					InstanceType: &instanceType,
				},
			},
		},
	)
}

func TestResources_JSON(t *testing.T) {
	JSON(
		t,
		reflect.TypeOf(block.Resources{}),
		[]JSONTest{
			{
				Description: "MockResources",
				Bytes:       MockResourcesExpectedJSON(),
				ValuePtr:    MockResources(),
			},
		},
	)
}
