package v1

import (
	"fmt"

	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	v1rest "github.com/mlab-lattice/lattice/pkg/api/v1/rest"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
)

const serviceIdentifier = "service_id"

var serviceIdentifierPathComponent = fmt.Sprintf(":%v", serviceIdentifier)

var servicesPath = fmt.Sprintf(v1rest.ServicesPathFormat, systemIdentifierPathComponent)
var servicePath = fmt.Sprintf(v1rest.ServicePathFormat, systemIdentifierPathComponent, serviceIdentifierPathComponent)
var serviceLogPath = fmt.Sprintf(v1rest.ServiceLogsPathFormat, systemIdentifierPathComponent,
	serviceIdentifierPathComponent)

func (api *LatticeAPI) setupServicesEndpoints() {
	// list-services
	api.router.GET(servicesPath, api.handleListServices)

	// get-service
	api.router.GET(servicePath, api.handleGetService)

	// service component log path
	api.router.GET(serviceLogPath, api.handleGetServiceLogs)

}

// ListServices godoc
// @ID list-services
// @Summary Lists services
// @Description list services
// @Router /v1/systems/{systemId}/services [get]
// @Param systemId path string true "System ID"
// @Accept  json
// @Produce  json
// @Success 200 {array} v1.Service
func (api *LatticeAPI) handleListServices(c *gin.Context) {
	systemID := v1.SystemID(c.Param(systemIdentifier))
	servicePathParam := c.Query("path")

	// check if its a query by service path

	if servicePathParam != "" {
		servicePath, err := tree.NewNodePath(servicePathParam)
		if err != nil {
			handleError(c, err)
			return
		}

		service, err := api.backend.GetServiceByPath(systemID, servicePath)

		if err != nil {
			handleError(c, err)
			return
		}

		if service == nil {
			c.Status(http.StatusBadRequest)
			return
		}

		c.JSON(http.StatusOK, []*v1.Service{service})
		return
	}

	// otherwise its just a normal list services request
	services, err := api.backend.ListServices(systemID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, services)
}

// handleGetService godoc
// @ID get-service
// @Summary Get service
// @Description get service
// @Router /v1/systems/{systemId}/services/{id} [get]
// @Param systemId path string true "System ID"
// @Param id path string true "Service ID"
// @Accept  json
// @Produce  json
// @Success 200 {object} v1.Service
func (api *LatticeAPI) handleGetService(c *gin.Context) {
	systemID := v1.SystemID(c.Param(systemIdentifier))
	serviceID := v1.ServiceID(c.Param(serviceIdentifier))

	service, err := api.backend.GetService(systemID, serviceID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, service)
}

// GetServiceLogs godoc
// @ID get-service-logs
// @Summary Get service logs
// @Description get service logs
// @Router /v1/systems/{systemId}/services/{id}/logs  [get]
// @Param systemId path string true "System ID"
// @Param id path string true "Service ID"
// @Param instance query string true "Instance"
// @Param sidecar query string false "Sidecar"
// @Param follow query string bool "Follow"
// @Param previous query boolean false "Previous"
// @Param timestamps query boolean false "Timestamps"
// @Param tail query integer false "tail"
// @Param since query string false "Since"
// @Param sinceTime query string false "Since Time"
// @Accept  json
// @Produce  json
// @Success 200 {string} string "log stream"
func (api *LatticeAPI) handleGetServiceLogs(c *gin.Context) {
	systemID := v1.SystemID(c.Param(systemIdentifier))
	serviceId := v1.ServiceID(c.Param(serviceIdentifier))
	instance := c.Query("instance")

	sidecarQuery, sidecarSet := c.GetQuery("sidecar")
	var sidecar *string
	if sidecarSet {
		sidecar = &sidecarQuery
	}

	logOptions, err := requestedLogOptions(c)

	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	log, err := api.backend.ServiceLogs(systemID, serviceId, sidecar, instance, logOptions)

	if err != nil {
		handleError(c, err)
		return
	}

	if log == nil {
		c.Status(http.StatusOK)
		return
	}

	serveLogFile(log, c)
}
