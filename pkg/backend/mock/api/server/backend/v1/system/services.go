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

// Services
func (b *ServiceBackend) List() ([]v1.Service, error) {
	b.backend.registry.Lock()
	defer b.backend.registry.Unlock()

	record, err := b.backend.systemRecordInitialized(b.systemID)
	if err != nil {
		return nil, err
	}

	var services []v1.Service
	for _, service := range record.Services {
		services = append(services, *(service.Service))
	}

	return services, nil
}

func (b *ServiceBackend) Get(id v1.ServiceID) (*v1.Service, error) {
	b.backend.registry.Lock()
	defer b.backend.registry.Unlock()

	record, err := b.backend.systemRecordInitialized(b.systemID)
	if err != nil {
		return nil, err
	}

	service, ok := record.Services[id]
	if !ok {
		return nil, v1.NewInvalidServiceIDError()
	}

	result := new(v1.Service)
	*result = *(service.Service)

	return result, nil
}

func (b *ServiceBackend) GetByPath(path tree.Path) (*v1.Service, error) {
	b.backend.registry.Lock()
	defer b.backend.registry.Unlock()

	record, err := b.backend.systemRecordInitialized(b.systemID)
	if err != nil {
		return nil, err
	}

	id, ok := record.ServicePaths[path]
	if !ok {
		return nil, v1.NewInvalidPathError()
	}

	service := record.Services[id]

	result := new(v1.Service)
	*result = *(service.Service)

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
