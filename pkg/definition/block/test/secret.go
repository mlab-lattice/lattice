package test

import (
	"github.com/mlab-lattice/system/pkg/definition/block"
	jsonutil "github.com/mlab-lattice/system/pkg/util/json"
)

func MockSecret() *block.Secret {
	secretName := "my-secret"
	return &block.Secret{
		Name: &secretName,
	}
}

func MockSecretReference() *block.Secret {
	return &block.Secret{
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
