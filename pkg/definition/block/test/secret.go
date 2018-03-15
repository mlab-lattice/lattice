package test

import (
	"github.com/mlab-lattice/system/pkg/definition/block"
	jsonutil "github.com/mlab-lattice/system/pkg/util/json"
)

func MockSecret() *block.SecretValue {
	secretName := "my-secret"
	return &block.SecretValue{
		Name: &secretName,
	}
}

func MockSecretReference() *block.SecretValue {
	return &block.SecretValue{
		Reference: MockReference(),
	}
}

func MockSecretExpectedJSON() []byte {
	return GenerateSecretExpectedJSON([]byte("\"my-secret\""), []byte(""))
}

func MockSecretReferenceExpectedJSON() []byte {
	return GenerateSecretExpectedJSON([]byte(""), MockReferenceExpectedJSON())
}

func GenerateSecretExpectedJSON(name, reference []byte) []byte {
	if string(name) != "" {
		return jsonutil.GenerateObjectBytes(
			[]jsonutil.FieldBytes{
				{
					Name:  "secret",
					Bytes: name,
				},
			},
		)
	}

	return jsonutil.GenerateObjectBytes([]jsonutil.FieldBytes{
		{
			Name:  "secret",
			Bytes: reference,
		},
	})
}
