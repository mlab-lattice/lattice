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
	EnvoyEntityType EntityType = iota
	LatticeEntityType
)

func (t EntityType) String() string {
	var _type string
	switch t {
	case EnvoyEntityType:
		_type = "EnvoyEntityType"
	case LatticeEntityType:
		_type = "LatticeEntityType"
	}
	return _type
}

type Event int

const (
	InformerAddEvent Event = iota
	InformerUpdateEvent
	InformerDeleteEvent

	EnvoyStreamRequestEvent
)

func (e Event) String() string {
	var event string
	switch e {
	case InformerAddEvent:
		event = "InformerAddEvent"
	case InformerUpdateEvent:
		event = "InformerUpdateEvent"
	case InformerDeleteEvent:
		event = "InformerDeleteEvent"
	case EnvoyStreamRequestEvent:
		event = "EnvoyStreamRequestEvent"
	}
	return event
}

type CacheUpdateTask struct {
	Name  string     `json:"name"`
	Type  EntityType `json:"type"`
	Event Event      `json:"event"`
}
