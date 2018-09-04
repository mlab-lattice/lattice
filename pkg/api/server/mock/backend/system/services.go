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
	b.backend.Lock()
	defer b.backend.Unlock()

	record, err := b.backend.systemRecord(b.systemID)
	if err != nil {
		return nil, err
	}

	var services []v1.Service
	for _, service := range record.system.Services {
		services = append(services, service)
	}

	return services, nil
}

func (b *ServiceBackend) Get(id v1.ServiceID) (*v1.Service, error) {
	b.backend.Lock()
	defer b.backend.Unlock()

	record, err := b.backend.systemRecord(b.systemID)
	if err != nil {
		return nil, err
	}

	for _, service := range record.system.Services {
		if service.ID == id {
			result := new(v1.Service)
			*result = service
			return result, nil
		}
	}

	return nil, v1.NewInvalidServiceIDError()
}

func (b *ServiceBackend) GetByPath(path tree.Path) (*v1.Service, error) {
	b.backend.Lock()
	defer b.backend.Unlock()

	record, err := b.backend.systemRecord(b.systemID)
	if err != nil {
		return nil, err
	}

	service, ok := record.system.Services[path]
	if !ok {
		return nil, v1.NewInvalidServicePathError()
	}

	result := new(v1.Service)
	*result = service

	return result, nil
}

func (b *ServiceBackend) Logs(
	id v1.ServiceID,
	sidecar *string,
	instance string,
	logOptions *v1.ContainerLogOptions,
) (io.ReadCloser, error) {
	_, err := b.Get(id)
	if err != nil {
		return nil, err
	}

	return ioutil.NopCloser(strings.NewReader("this is a long line")), nil
}
