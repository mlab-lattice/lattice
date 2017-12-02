package test

import (
	"github.com/mlab-lattice/system/pkg/definition/block"
	jsonutil "github.com/mlab-lattice/system/pkg/util/json"
)

func MockPort() *block.ComponentPort {
	return MockHttpPort()
}

func MockPortExpectedJson() []byte {
	return MockHttpPortExpectedJson()
}

func MockHttpPort() *block.ComponentPort {
	return MockPrivateHttpPort()
}

func MockHttpPortExpectedJson() []byte {
	return MockPrivateHttpPortExpectedJson()
}

func MockPrivateHttpPort() *block.ComponentPort {
	return &block.ComponentPort{
		Name:     "http",
		Port:     80,
		Protocol: block.HttpProtocol,
	}
}

func MockPrivateHttpPortExpectedJson() []byte {
	httpProtocolBytes := []byte(`"`)
	httpProtocolBytes = append(httpProtocolBytes, []byte(block.HttpProtocol)...)
	httpProtocolBytes = append(httpProtocolBytes, []byte(`"`)...)
	return GeneratePortExpectedJson(
		[]byte(`"http"`),
		[]byte(`80`),
		httpProtocolBytes,
		nil,
	)
}

func MockPublicHttpPort() *block.ComponentPort {
	return &block.ComponentPort{
		Name:           "http",
		Port:           80,
		Protocol:       block.HttpProtocol,
		ExternalAccess: MockPublicExternalAccess(),
	}
}

func MockPublicHttpPortExpectedJson() []byte {
	httpProtocolBytes := []byte(`"`)
	httpProtocolBytes = append(httpProtocolBytes, []byte(block.HttpProtocol)...)
	httpProtocolBytes = append(httpProtocolBytes, []byte(`"`)...)
	return GeneratePortExpectedJson(
		[]byte(`"http"`),
		[]byte(`80`),
		httpProtocolBytes,
		MockPublicExternalAccessExpectedJson(),
	)
}

func MockTcpPort() *block.ComponentPort {
	return MockPrivateTcpPort()
}

func MockTcpPortExpectedJson() []byte {
	return MockPrivateTcpPortExpectedJson()
}

func MockPrivateTcpPort() *block.ComponentPort {
	return &block.ComponentPort{
		Name:     "tcp",
		Port:     80,
		Protocol: block.TcpProtocol,
	}
}

func MockPrivateTcpPortExpectedJson() []byte {
	tcpProtocolBytes := []byte(`"`)
	tcpProtocolBytes = append(tcpProtocolBytes, []byte(block.TcpProtocol)...)
	tcpProtocolBytes = append(tcpProtocolBytes, []byte(`"`)...)
	return GeneratePortExpectedJson(
		[]byte(`"tcp"`),
		[]byte(`80`),
		tcpProtocolBytes,
		nil,
	)
}

func MockPublicTcpPort() *block.ComponentPort {
	return &block.ComponentPort{
		Name:           "tcp",
		Port:           80,
		Protocol:       block.TcpProtocol,
		ExternalAccess: MockPublicExternalAccess(),
	}
}

func MockPublicTcpPortExpectedJson() []byte {
	tcpProtocolBytes := []byte(`"`)
	tcpProtocolBytes = append(tcpProtocolBytes, []byte(block.TcpProtocol)...)
	tcpProtocolBytes = append(tcpProtocolBytes, []byte(`"`)...)
	return GeneratePortExpectedJson(
		[]byte(`"tcp"`),
		[]byte(`80`),
		tcpProtocolBytes,
		MockPublicExternalAccessExpectedJson(),
	)
}

func GeneratePortExpectedJson(name, port, protocol, externalAccess []byte) []byte {
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

func MockExternalAccessExpectedJson() []byte {
	return MockPublicExternalAccessExpectedJson()
}

func MockPublicExternalAccess() *block.ExternalAccess {
	return &block.ExternalAccess{
		Public: true,
	}
}

func MockPublicExternalAccessExpectedJson() []byte {
	return GenerateExternalAccessExpectedJson([]byte(`true`))
}

func GenerateExternalAccessExpectedJson(public []byte) []byte {
	return jsonutil.GenerateObjectBytes([]jsonutil.FieldBytes{
		{
			Name:  "public",
			Bytes: public,
		},
	})
}
