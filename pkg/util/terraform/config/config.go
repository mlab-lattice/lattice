package config

type Config struct {
	Provider interface{}       `json:"provider,omitempty"`
	Modules  map[string]Module `json:"module,omitempty"`
}
