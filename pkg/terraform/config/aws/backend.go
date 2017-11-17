package aws

import (
	"encoding/json"
)

type S3Backend struct {
	Bucket  string
	Key     string
	Encrypt bool
}

// Implement json.Marshaler
func (b S3Backend) MarshalJSON() ([]byte, error) {
	moduleMap := map[string]interface{}{
		"s3": map[string]interface{}{
			"bucket":  b.Bucket,
			"key":     b.Key,
			"encrypt": b.Encrypt,
		},
	}
	return json.Marshal(moduleMap)
}
