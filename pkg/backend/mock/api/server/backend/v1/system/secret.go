package system

import (
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
)

type SecretBackend struct {
	systemID v1.SystemID
	backend  *Backend
}

// Secrets
func (b *SecretBackend) List() ([]v1.Secret, error) {
	b.backend.Lock()
	defer b.backend.Unlock()

	record, err := b.backend.systemRecord(b.systemID)
	if err != nil {
		return nil, err
	}

	var secrets []v1.Secret
	for _, secret := range record.secrets {
		secrets = append(secrets, *secret)
	}

	return secrets, nil
}

func (b *SecretBackend) Get(path tree.PathSubcomponent) (*v1.Secret, error) {
	b.backend.Lock()
	defer b.backend.Unlock()

	record, err := b.backend.systemRecord(b.systemID)
	if err != nil {
		return nil, err
	}

	for _, secret := range record.secrets {
		if secret.Path == path {
			result := new(v1.Secret)
			*result = *secret
			return result, nil
		}
	}

	return nil, v1.NewInvalidSecretError()
}

func (b *SecretBackend) Set(path tree.PathSubcomponent, value string) error {
	b.backend.Lock()
	defer b.backend.Unlock()

	record, err := b.backend.systemRecord(b.systemID)
	if err != nil {
		return err
	}

	for _, secret := range record.secrets {
		if secret.Path == path {
			secret.Value = value
			return nil
		}
	}

	secret := &v1.Secret{
		Path:  path,
		Value: value,
	}

	record.secrets = append(record.secrets, secret)

	return nil
}

func (b *SecretBackend) Unset(path tree.PathSubcomponent) error {
	b.backend.Lock()
	defer b.backend.Unlock()

	record, err := b.backend.systemRecord(b.systemID)
	if err != nil {
		return err
	}

	for i, secret := range record.secrets {
		if secret.Path == path {
			// delete secret
			record.secrets = append(record.secrets[:i], record.secrets[i+1:]...)

			return nil
		}
	}

	return v1.NewInvalidSecretError()
}
