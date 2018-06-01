package v2

type Service struct {
	EgressPort  int32
	Components  map[string]Component
	IPAddresses []string
}

type Component struct {
	// Ports maps the Component's ports to their envoy ports.
	Ports map[int32]int32
}

type EntityType int

const (
	KubeEntityType EntityType = iota
	LatticeEntityType
)

func (t EntityType) String() string {
	var _type string
	switch t {
	case KubeEntityType:
		_type = "KubeEntityType"
	case LatticeEntityType:
		_type = "LatticeEntityType"
	}
	return _type
}

type InformerEvent int

const (
	InformerAddEvent InformerEvent = iota
	InformerUpdateEvent
	InformerDeleteEvent
)

func (e InformerEvent) String() string {
	var event string
	switch e {
	case InformerAddEvent:
		event = "InformerAddEvent"
	case InformerUpdateEvent:
		event = "InformerUpdateEvent"
	case InformerDeleteEvent:
		event = "InformerDeleteEvent"
	}
	return event
}

type CacheUpdateTask struct {
	Name string     `json:"name"`
	Type EntityType `json:"type"`
}
