package block

import (
	"errors"
	"fmt"
	"strings"
)

// TODO: review this limit
const MaxVolumeSizeInGb = 1024

type Volume struct {
	Name     string `json:"name"`
	SizeInGb uint32 `json:"size_in_gb"`
}

// Validate implements Interface
func (v *Volume) Validate(interface{}) error {
	if v.Name == "" {
		return errors.New("name is required")
	}

	if v.SizeInGb < 1 {
		return errors.New("invalid size_in_gb")
	}

	if v.SizeInGb > MaxVolumeSizeInGb {
		return fmt.Errorf("size_in_gb %v exceeds maximum %v", v.SizeInGb, MaxVolumeSizeInGb)
	}

	return nil
}

type ComponentVolumeMount struct {
	Name       string `json:"name"`
	MountPoint string `json:"mount_point"`
	ReadOnly   bool   `json:"read_only"`
}

// Validate implements Interface
func (v *ComponentVolumeMount) Validate(interface{}) error {
	if v.Name == "" {
		return errors.New("name is required")
	}

	if !strings.HasPrefix(v.MountPoint, "/") {
		return errors.New("mount_path must begin with '/'")
	}

	return nil
}
