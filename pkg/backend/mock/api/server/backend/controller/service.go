package controller

import (
	"log"
	"math"
	"sync"
	"time"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/backend/mock/api/server/backend/registry"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	definitionv1 "github.com/mlab-lattice/lattice/pkg/definition/v1"

	"github.com/satori/go.uuid"
)

const (
	serviceScaleRate   = 0.3
	serviceScalePeriod = 2 * time.Second
)

func (c *Controller) addService(
	path tree.Path,
	definition *definitionv1.Service,
	record *registry.SystemRecord,
	wg *sync.WaitGroup,
) {
	log.Printf("adding service %v for system %v", path.String(), record.System.ID)

	defer wg.Done()

	var service *v1.Service

	func() {
		c.registry.Lock()
		defer c.registry.Unlock()

		service = &v1.Service{
			ID:   v1.ServiceID(uuid.NewV4().String()),
			Path: path,

			State: v1.ServiceStatePending,

			AvailableInstances:   0,
			UpdatedInstances:     0,
			StaleInstances:       0,
			TerminatingInstances: 0,

			Ports: make(map[int32]string),

			Instances: make([]string, 0),
		}

		record.Services[service.ID] = &registry.ServiceInfo{
			Service:    service,
			Definition: definition,
		}

		record.ServicePaths[path] = service.ID
	}()

	for {
		time.Sleep(serviceScalePeriod)

		done := func() bool {
			c.registry.Lock()
			defer c.registry.Unlock()

			log.Printf("scaling service %v for system %v", path.String(), record.System.ID)

			// increase the total number of instances available as a factor of the desired number
			available := int32(math.Min(
				math.Ceil(float64((1+serviceScaleRate)*float64(definition.NumInstances))),
				float64(definition.NumInstances),
			))

			service.AvailableInstances = available
			service.UpdatedInstances = available

			// add new instance ids
			diff := available - service.AvailableInstances
			var newInstances []string
			for i := int32(0); i < diff; i++ {
				newInstances = append(newInstances, uuid.NewV4().String())
			}

			service.Instances = append(service.Instances, newInstances...)

			// if we've reached the desired number of instances, we're done
			return service.AvailableInstances == definition.NumInstances
		}()
		if done {
			c.registry.Lock()
			defer c.registry.Unlock()

			service.State = v1.ServiceStateStable

			log.Printf("done scaling service %v for system %v", path.String(), record.System.ID)
			return
		}
	}
}

func (c *Controller) rollService(
	path tree.Path,
	definition *definitionv1.Service,
	record *registry.SystemRecord,
	wg *sync.WaitGroup,
) {
	log.Printf("beginning rolling scaling service %v for system %v", path.String(), record.System.ID)

	defer wg.Done()

	var service *v1.Service
	func() {
		c.registry.Lock()
		defer c.registry.Unlock()

		id := record.ServicePaths[path]
		service = record.Services[id].Service
		service.State = v1.ServiceStateUpdating
		service.StaleInstances = service.AvailableInstances
		service.UpdatedInstances = 0
	}()

	desired := definition.NumInstances
	var newInstances []string
	for i := int32(0); i < desired; i++ {
		newInstances = append(newInstances, uuid.NewV4().String())
	}

	var oldInstances []string
	for _, i := range service.Instances {
		oldInstances = append(oldInstances, i)
	}

	for {
		time.Sleep(serviceScalePeriod)

		done := func() bool {
			c.registry.Lock()
			defer c.registry.Unlock()

			log.Printf("rolling scaling service %v for system %v", path.String(), record.System.ID)

			rev := int32(math.Ceil(float64(definition.NumInstances) * serviceScaleRate))
			service.UpdatedInstances = int32(math.Min(float64(service.UpdatedInstances+rev), float64(desired)))
			if service.StaleInstances == 0 {
				service.TerminatingInstances = int32(math.Max(float64(service.TerminatingInstances-rev), 0))
			} else {
				service.StaleInstances = int32(math.Max(float64(service.TerminatingInstances-rev), 0))
			}

			service.AvailableInstances = service.UpdatedInstances + service.StaleInstances
			service.Instances = append(
				oldInstances[:(service.StaleInstances+service.TerminatingInstances)],
				newInstances[:service.UpdatedInstances]...,
			)

			return service.UpdatedInstances == desired
		}()
		if done {
			c.registry.Lock()
			defer c.registry.Unlock()

			service.State = v1.ServiceStateStable

			log.Printf("done rolling service %v for system %v", path.String(), record.System.ID)
			return
		}
	}
}

func (c *Controller) terminateService(path tree.Path, record *registry.SystemRecord, wg *sync.WaitGroup) {
	defer wg.Done()

	log.Printf("beginning terminating service %v for system %v", path.String(), record.System.ID)

	var service *v1.Service
	func() {
		c.registry.Lock()
		defer c.registry.Unlock()

		id := record.ServicePaths[path]
		service = record.Services[id].Service
		service.State = v1.ServiceStateDeleting
	}()

	for {
		time.Sleep(serviceScalePeriod)

		done := func() bool {
			c.registry.Lock()
			defer c.registry.Unlock()

			log.Printf("terminating service %v for system %v", path.String(), record.System.ID)

			removeTerminating := int32(math.Ceil(float64(service.TerminatingInstances) * serviceScaleRate))
			remainingTerminating := service.TerminatingInstances - removeTerminating + service.AvailableInstances + service.StaleInstances
			service.AvailableInstances = 0
			service.UpdatedInstances = 0
			service.StaleInstances = 0
			service.TerminatingInstances = remainingTerminating
			service.Instances = service.Instances[:remainingTerminating]

			return service.TerminatingInstances == 0
		}()
		if done {
			c.registry.Lock()
			defer c.registry.Unlock()

			log.Printf("done terminating service %v for system %v", path.String(), record.System.ID)

			delete(record.Services, service.ID)
			delete(record.ServicePaths, path)

			return
		}
	}
}
