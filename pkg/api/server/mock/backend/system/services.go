package system

import (
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	"io"
	"io/ioutil"
	"strings"
)

type ServiceBackend struct {
	systemID v1.SystemID
	backend  *Backend
}

// Secrets
func (b *ServiceBackend) List() ([]v1.Service, error) {
	b.backend.registryLock.RLock()
	defer b.backend.registryLock.RUnlock()

	record, err := b.backend.getSystemRecordLocked(b.systemID)
	if err != nil {
		return nil, err
	}

	record.recordLock.RLock()
	defer record.recordLock.RUnlock()

	var services []v1.Service
	for _, service := range record.system.Services {
		services = append(services, service)
	}

	return services, nil
}

func (b *ServiceBackend) Get(id v1.ServiceID) (*v1.Service, error) {
	b.backend.registryLock.RLock()
	defer b.backend.registryLock.RUnlock()

	record, err := b.backend.getSystemRecordLocked(b.systemID)
	if err != nil {
		return nil, err
	}

	record.recordLock.RLock()
	defer record.recordLock.RUnlock()

	for _, service := range record.system.Services {
		if service.ID == id {
			return &service, nil
		}
	}

	return nil, v1.NewInvalidServiceIDError(id)
}

func (b *ServiceBackend) GetByPath(path tree.Path) (*v1.Service, error) {
	b.backend.registryLock.RLock()
	defer b.backend.registryLock.RUnlock()

	record, err := b.backend.getSystemRecordLocked(b.systemID)
	if err != nil {
		return nil, err
	}

	record.recordLock.RLock()
	defer record.recordLock.RUnlock()

	service, ok := record.system.Services[path]
	if !ok {
		return nil, v1.NewInvalidServicePathError(path)
	}

	return &service, nil
}

func (b *ServiceBackend) Logs(
	id v1.ServiceID,
	sidecar *string,
	instance string,
	logOptions *v1.ContainerLogOptions,
) (io.ReadCloser, error) {
	return ioutil.NopCloser(strings.NewReader("this is a long line")), nil
}
