package envoy

import (
	"net"
)

type ProtoToCIDRBlock struct {
	HTTP net.IPNet
	TCP  net.IPNet
}
