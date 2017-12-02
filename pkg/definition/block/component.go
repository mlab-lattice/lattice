package block

import (
	"errors"
	"fmt"
)

type Component struct {
	Name         string                  `json:"name"`
	Init         bool                    `json:"init"`
	Ports        []*ComponentPort        `json:"ports,omitempty"`
	VolumeMounts []*ComponentVolumeMount `json:"volume_mounts,omitempty"`
	Build        ComponentBuild          `json:"build"`
	Exec         ComponentExec           `json:"exec"`
	HealthCheck  *ComponentHealthCheck   `json:"health_check,omitempty"`
}

// Implement Interface
func (c *Component) Validate(information interface{}) error {
	if c.Name == "" {
		return errors.New("name is required")
	}

	volumes := information.(map[string]*Volume)

	ports := map[string]*ComponentPort{}
	for _, port := range c.Ports {
		if err := port.Validate(nil); err != nil {
			return fmt.Errorf("port %v definition error: %v", port.Name, err)
		}
		ports[port.Name] = port
	}

	for _, volumeMount := range c.VolumeMounts {
		if _, exists := volumes[volumeMount.Name]; !exists {
			return fmt.Errorf("invalid volume name %v in volume_mounts", volumeMount.Name)
		}

		if err := volumeMount.Validate(nil); err != nil {
			return fmt.Errorf("volume_mount %v definition error: %v", volumeMount.Name, err)
		}
	}

	if err := c.Build.Validate(nil); err != nil {
		return fmt.Errorf("build definition error: %v", err)
	}

	if err := c.Exec.Validate(nil); err != nil {
		return fmt.Errorf("exec definition error: %v", err)
	}

	if c.HealthCheck != nil {
		if err := c.HealthCheck.Validate(ports); err != nil {
			return fmt.Errorf("health_check definition error: %v", err)
		}
	}

	return nil
}
