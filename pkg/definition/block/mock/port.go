package mock

import (
	"github.com/mlab-lattice/system/pkg/definition/block"
	jsonutil "github.com/mlab-lattice/system/pkg/util/json"
)

func Port() *block.ComponentPort {
	return HTTPPort()
}

func PortExpectedJSON() []byte {
	return HTTPPortExpectedJSON()
}

func HTTPPort() *block.ComponentPort {
	return PrivateHTTPPort()
}

func HTTPPortExpectedJSON() []byte {
	return PrivateHTTPPortExpectedJSON()
}

func PrivateHTTPPort() *block.ComponentPort {
	return &block.ComponentPort{
		Name:     "http",
		Port:     80,
		Protocol: block.ProtocolHTTP,
	}
}

func PrivateHTTPPortExpectedJSON() []byte {
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

func PublicHTTPPort() *block.ComponentPort {
	return &block.ComponentPort{
		Name:           "http",
		Port:           80,
		Protocol:       block.ProtocolHTTP,
		ExternalAccess: PublicExternalAccess(),
	}
}

func PublicHTTPPortExpectedJSON() []byte {
	httpProtocolBytes := []byte(`"`)
	httpProtocolBytes = append(httpProtocolBytes, []byte(block.ProtocolHTTP)...)
	httpProtocolBytes = append(httpProtocolBytes, []byte(`"`)...)
	return GeneratePortExpectedJSON(
		[]byte(`"http"`),
		[]byte(`80`),
		httpProtocolBytes,
		PublicExternalAccessExpectedJSON(),
	)
}

func TCPPort() *block.ComponentPort {
	return PrivateTCPPort()
}

func TCPPortExpectedJSON() []byte {
	return PrivateTCPPortExpectedJSON()
}

func PrivateTCPPort() *block.ComponentPort {
	return &block.ComponentPort{
		Name:     "tcp",
		Port:     80,
		Protocol: block.ProtocolTCP,
	}
}

func PrivateTCPPortExpectedJSON() []byte {
	tcpProtocolBytes := []byte(`"`)
	tcpProtocolBytes = append(tcpProtocolBytes, []byte(block.ProtocolTCP)...)
	tcpProtocolBytes = append(tcpProtocolBytes, []byte(`"`)...)
	return GeneratePortExpectedJSON(
		[]byte(`"tcp"`),
		[]byte(`80`),
		tcpProtocolBytes,
		nil,
	)
}

func PublicTCPPort() *block.ComponentPort {
	return &block.ComponentPort{
		Name:           "tcp",
		Port:           80,
		Protocol:       block.ProtocolTCP,
		ExternalAccess: PublicExternalAccess(),
	}
}

func PublicTCPPortExpectedJSON() []byte {
	tcpProtocolBytes := []byte(`"`)
	tcpProtocolBytes = append(tcpProtocolBytes, []byte(block.ProtocolTCP)...)
	tcpProtocolBytes = append(tcpProtocolBytes, []byte(`"`)...)
	return GeneratePortExpectedJSON(
		[]byte(`"tcp"`),
		[]byte(`80`),
		tcpProtocolBytes,
		PublicExternalAccessExpectedJSON(),
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

func ExternalAccess() *block.ExternalAccess {
	return PublicExternalAccess()
}

func ExternalAccessExpectedJSON() []byte {
	return PublicExternalAccessExpectedJSON()
}

func PublicExternalAccess() *block.ExternalAccess {
	return &block.ExternalAccess{
		Public: true,
	}
}

func PublicExternalAccessExpectedJSON() []byte {
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
