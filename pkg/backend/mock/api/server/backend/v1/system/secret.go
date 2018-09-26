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
	b.backend.registry.Lock()
	defer b.backend.registry.Unlock()

	record, err := b.backend.systemRecordInitialized(b.systemID)
	if err != nil {
		return nil, err
	}

	var secrets []v1.Secret
	for _, secret := range record.Secrets {
		secrets = append(secrets, *secret.DeepCopy())
	}

	return secrets, nil
}

func (b *SecretBackend) Get(path tree.PathSubcomponent) (*v1.Secret, error) {
	b.backend.registry.Lock()
	defer b.backend.registry.Unlock()

	record, err := b.backend.systemRecordInitialized(b.systemID)
	if err != nil {
		return nil, err
	}

	for _, secret := range record.Secrets {
		if secret.Path == path {
			return secret.DeepCopy(), nil
		}
	}

	return nil, v1.NewInvalidSecretError()
}

func (b *SecretBackend) Set(path tree.PathSubcomponent, value string) error {
	b.backend.registry.Lock()
	defer b.backend.registry.Unlock()

	record, err := b.backend.systemRecordInitialized(b.systemID)
	if err != nil {
		return err
	}

	for _, secret := range record.Secrets {
		if secret.Path == path {
			secret.Value = value
			return nil
		}
	}

	secret := &v1.Secret{
		Path:  path,
		Value: value,
	}

	record.Secrets[path] = secret

	return nil
}

func (b *SecretBackend) Unset(path tree.PathSubcomponent) error {
	b.backend.registry.Lock()
	defer b.backend.registry.Unlock()

	record, err := b.backend.systemRecordInitialized(b.systemID)
	if err != nil {
		return err
	}

	_, ok := record.Secrets[path]
	if !ok {
		return v1.NewInvalidSecretError()
	}

	delete(record.Secrets, path)
	return nil
}
