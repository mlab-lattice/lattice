package terraform

import (
	"encoding/json"
	"fmt"

	"github.com/mlab-lattice/lattice/pkg/util/cli2"
	"github.com/mlab-lattice/lattice/pkg/util/cli2/flags"
)

const (
	BackendS3 = "S3"
)

type BackendOptions struct {
	S3 *BackendOptionsS3
}

func (o *BackendOptions) AsFlags() []string {
	if o.S3 != nil {
		return o.S3.AsFlags()
	}

	return nil
}

type BackendOptionsS3 struct {
	Bucket string
}

func (o *BackendOptionsS3) AsFlags() []string {
	return []string{
		"terraform-backend=S3",
		fmt.Sprintf("terraform-backend-var=bucket=%v", o.Bucket),
	}
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

func BackendS3Flags() (cli.Flags, *BackendOptionsS3) {
	options := &BackendOptionsS3{}
	flags := cli.Flags{
		"bucket": &flags.String{
			Required: true,
			Target:   &options.Bucket,
		},
	}
	return flags, options
}

func BackendFlags(backend *string) (cli.Flag, *BackendOptions) {
	s3Flags, s3Options := BackendS3Flags()
	options := &BackendOptions{}

	flag := &flags.DelayedEmbedded{
		Required: false,
		Usage:    "configuration for the terraform backend",
		Flags: map[string]cli.Flags{
			BackendS3: s3Flags,
		},
		FlagChooser: func() (*string, error) {
			if backend == nil || *backend == "" {
				return nil, nil
			}

			switch *backend {
			case BackendS3:
				options.S3 = s3Options
			default:
				return nil, fmt.Errorf("unsupported terraform backend %v", *backend)
			}

			return backend, nil
		},
	}

	return flag, options
}
