package terraform

import (
	"encoding/json"
)

const (
	BackendS3 = "S3"
)

type BackendOptions struct {
	S3 *BackendOptionsS3
}

type BackendOptionsS3 struct {
	Bucket string
}

type S3BackendConfig struct {
	Region  string
	Bucket  string
	Key     string
	Encrypt bool
}

// MarshalJSON implements json.Marshaler
func (b S3BackendConfig) MarshalJSON() ([]byte, error) {
	moduleMap := map[string]interface{}{
		"s3": map[string]interface{}{
			"region":  b.Region,
			"bucket":  b.Bucket,
			"key":     b.Key,
			"encrypt": b.Encrypt,
		},
	}
	return json.Marshal(moduleMap)
}
