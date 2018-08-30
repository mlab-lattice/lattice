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
	b.backend.registryLock.RLock()
	defer b.backend.registryLock.RUnlock()

	record, err := b.backend.getSystemRecordLocked(b.systemID)
	if err != nil {
		return nil, err
	}

	record.recordLock.RLock()
	defer record.recordLock.RUnlock()

	return record.secrets, nil
}

func (b *SecretBackend) Get(path tree.PathSubcomponent) (*v1.Secret, error) {
	b.backend.registryLock.RLock()
	defer b.backend.registryLock.RUnlock()

	record, err := b.backend.getSystemRecordLocked(b.systemID)
	if err != nil {
		return nil, err
	}

	record.recordLock.RLock()
	defer record.recordLock.RUnlock()

	for _, secret := range record.secrets {
		if secret.Path == path {
			return &secret, nil
		}
	}

	return nil, v1.NewInvalidSystemSecretError(path)
}

func (b *SecretBackend) Set(path tree.PathSubcomponent, value string) error {
	b.backend.registryLock.RLock()
	defer b.backend.registryLock.RUnlock()

	record, err := b.backend.getSystemRecordLocked(b.systemID)
	if err != nil {
		return err
	}

	record.recordLock.Lock()
	defer record.recordLock.Unlock()

	for _, secret := range record.secrets {
		if secret.Path == path {
			secret.Value = value
			return nil
		}
	}

	secret := v1.Secret{
		Path:  path,
		Value: value,
	}

	record.secrets = append(record.secrets, secret)

	return nil
}

func (b *SecretBackend) Unset(path tree.PathSubcomponent) error {
	b.backend.registryLock.RLock()
	defer b.backend.registryLock.RUnlock()

	record, err := b.backend.getSystemRecordLocked(b.systemID)
	if err != nil {
		return err
	}

	record.recordLock.Lock()
	defer record.recordLock.Unlock()

	for i, secret := range record.secrets {
		if secret.Path == path {
			// delete secret
			record.secrets = append(record.secrets[:i], record.secrets[i+1:]...)

			return nil
		}
	}

	return v1.NewInvalidSystemSecretError(path)
}
