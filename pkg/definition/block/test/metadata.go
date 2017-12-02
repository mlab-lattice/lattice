package test

import (
	sd "github.com/mlab-lattice/system/pkg/definition"
	"github.com/mlab-lattice/system/pkg/definition/block"
	jsonutil "github.com/mlab-lattice/system/pkg/util/json"
)

func MockSystemMetadata() *block.Metadata {
	return &block.Metadata{
		Name:        "my-system",
		Type:        sd.SystemType,
		Description: "This is my system",
	}
}

func MockSystemMetadataExpectedJson() []byte {
	serviceTypeBytes := []byte(`"`)
	serviceTypeBytes = append(serviceTypeBytes, []byte(sd.SystemType)...)
	serviceTypeBytes = append(serviceTypeBytes, []byte(`"`)...)
	return GenerateMetadataExpectedJson(
		[]byte(`"my-system"`),
		serviceTypeBytes,
		[]byte(`"This is my system"`),
		nil,
	)
}

func MockServiceMetadata() *block.Metadata {
	return &block.Metadata{
		Name:        "my-service",
		Type:        sd.ServiceType,
		Description: "This is my service",
	}
}

func MockServiceMetadataExpectedJson() []byte {
	serviceTypeBytes := []byte(`"`)
	serviceTypeBytes = append(serviceTypeBytes, []byte(sd.ServiceType)...)
	serviceTypeBytes = append(serviceTypeBytes, []byte(`"`)...)
	return GenerateMetadataExpectedJson(
		[]byte(`"my-service"`),
		serviceTypeBytes,
		[]byte(`"This is my service"`),
		nil,
	)
}

func MockServiceDifferentNameMetadata() *block.Metadata {
	return &block.Metadata{
		Name:        "my-other-service",
		Type:        sd.ServiceType,
		Description: "This is my service",
	}
}

func MockServiceDifferentNameMetadataExpectedJson() []byte {
	serviceTypeBytes := []byte(`"`)
	serviceTypeBytes = append(serviceTypeBytes, []byte(sd.ServiceType)...)
	serviceTypeBytes = append(serviceTypeBytes, []byte(`"`)...)
	return GenerateMetadataExpectedJson(
		[]byte(`"my-other-service"`),
		serviceTypeBytes,
		[]byte(`"This is my service"`),
		nil,
	)
}

func GenerateMetadataExpectedJson(name, _type, description, parameters []byte) []byte {
	return jsonutil.GenerateObjectBytes([]jsonutil.FieldBytes{
		{
			Name:  "name",
			Bytes: name,
		},
		{
			Name:  "type",
			Bytes: _type,
		},
		{
			Name:  "description",
			Bytes: description,
		},
		{
			Name:      "parameters",
			Bytes:     parameters,
			OmitEmpty: true,
		},
	})
}
