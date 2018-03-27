package test

import (
	"github.com/mlab-lattice/system/pkg/definition/block"
	jsonutil "github.com/mlab-lattice/system/pkg/util/json"
)

func MockPort() *block.ComponentPort {
	return MockHTTPPort()
}

func MockPortExpectedJSON() []byte {
	return MockHTTPPortExpectedJSON()
}

func MockHTTPPort() *block.ComponentPort {
	return MockPrivateHTTPPort()
}

func MockHTTPPortExpectedJSON() []byte {
	return MockPrivateHTTPPortExpectedJSON()
}

func MockPrivateHTTPPort() *block.ComponentPort {
	return &block.ComponentPort{
		Name:     "http",
		Port:     80,
		Protocol: block.ProtocolHTTP,
	}
}

func MockPrivateHTTPPortExpectedJSON() []byte {
	httpProtocolBytes := []byte(`"`)
	httpProtocolBytes = append(httpProtocolBytes, []byte(block.ProtocolHTTP)...)
	httpProtocolBytes = append(httpProtocolBytes, []byte(`"`)...)
	return GeneratePortExpectedJSON(
		[]byte(`"http"`),
		[]byte(`80`),
		httpProtocolBytes,
		nil,
	)
}

func MockPublicHTTPPort() *block.ComponentPort {
	return &block.ComponentPort{
		Name:           "http",
		Port:           80,
		Protocol:       block.ProtocolHTTP,
		ExternalAccess: MockPublicExternalAccess(),
	}
}

func MockPublicHTTPPortExpectedJSON() []byte {
	httpProtocolBytes := []byte(`"`)
	httpProtocolBytes = append(httpProtocolBytes, []byte(block.ProtocolHTTP)...)
	httpProtocolBytes = append(httpProtocolBytes, []byte(`"`)...)
	return GeneratePortExpectedJSON(
		[]byte(`"http"`),
		[]byte(`80`),
		httpProtocolBytes,
		MockPublicExternalAccessExpectedJSON(),
	)
}

func GeneratePortExpectedJSON(name, port, protocol, externalAccess []byte) []byte {
	return jsonutil.GenerateObjectBytes([]jsonutil.FieldBytes{
		{
			Name:  "name",
			Bytes: name,
		},
		{
			Name:  "port",
			Bytes: port,
		},
		{
			Name:  "protocol",
			Bytes: protocol,
		},
		{
			Name:      "external_access",
			Bytes:     externalAccess,
			OmitEmpty: true,
		},
	})
}

func MockExternalAccess() *block.ExternalAccess {
	return MockPublicExternalAccess()
}

func MockExternalAccessExpectedJSON() []byte {
	return MockPublicExternalAccessExpectedJSON()
}

func MockPublicExternalAccess() *block.ExternalAccess {
	return &block.ExternalAccess{
		Public: true,
	}
}

func MockPublicExternalAccessExpectedJSON() []byte {
	return GenerateExternalAccessExpectedJSON([]byte(`true`))
}

func GenerateExternalAccessExpectedJSON(public []byte) []byte {
	return jsonutil.GenerateObjectBytes([]jsonutil.FieldBytes{
		{
			Name:  "public",
			Bytes: public,
		},
	})
}
