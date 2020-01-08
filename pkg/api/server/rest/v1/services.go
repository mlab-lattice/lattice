package v1

import (
	"fmt"

	"net/http"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	v1rest "github.com/mlab-lattice/lattice/pkg/api/v1/rest"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"

	"github.com/gin-gonic/gin"
)

const serviceIdentifier = "service_id"

var (
	serviceIdentifierPathComponent = fmt.Sprintf(":%v", serviceIdentifier)
	servicesPath                   = fmt.Sprintf(v1rest.ServicesPathFormat, systemIdentifierPathComponent)
	servicePath                    = fmt.Sprintf(v1rest.ServicePathFormat, systemIdentifierPathComponent, serviceIdentifierPathComponent)
	serviceLogPath                 = fmt.Sprintf(v1rest.ServiceLogsPathFormat, systemIdentifierPathComponent, serviceIdentifierPathComponent)
)

func (api *LatticeAPI) setupServicesEndpoints() {
	// list-services
	api.router.GET(servicesPath, api.handleListServices)

	// get-service
	api.router.GET(servicePath, api.handleGetService)

	// service component log path
	api.router.GET(serviceLogPath, api.handleGetServiceLogs)

}

// swagger:operation GET /systems/{system}/services services ListServices
//
// Lists services
//
// Lists services
// ---
//     consumes:
//     - application/json
//     produces:
//     - application/json
//
//     parameters:
//       - description: System ID
//         in: path
//         name: system
//         required: true
//         type: string
//
//     responses:
//         '200':
//           description: service list
//           schema:
//             type: array
//             items:
//               "$ref": "#/definitions/Service"
//

// handleListServices handler for ListServices
func (api *LatticeAPI) handleListServices(c *gin.Context) {
	systemID := v1.SystemID(c.Param(systemIdentifier))
	servicePathParam := c.Query("path")

	// check if its a query by service path

	if servicePathParam != "" {
		path, err := tree.NewPath(servicePathParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, v1.NewInvalidPathError())
			return
		}

		service, err := api.backend.Systems().Services(systemID).GetByPath(path)
		if err != nil {
			v1err, ok := err.(*v1.Error)
			if !ok {
				handleInternalError(c, err)
				return
			}

			switch v1err.Code {
			case v1.ErrorCodeInvalidSystemID:
				c.JSON(http.StatusNotFound, v1err)

			case v1.ErrorCodeInvalidPath:
				c.JSON(http.StatusOK, []*v1.Service{})

			default:
				handleInternalError(c, err)
			}
			return
		}

		c.JSON(http.StatusOK, []*v1.Service{service})
		return
	}

	// otherwise its just a normal list services request
	services, err := api.backend.Systems().Services(systemID).List()
	if err != nil {
		v1err, ok := err.(*v1.Error)
		if !ok {
			handleInternalError(c, err)
			return
		}

		switch v1err.Code {
		case v1.ErrorCodeInvalidSystemID:
			c.JSON(http.StatusNotFound, v1err)

		default:
			handleInternalError(c, err)
		}
		return
	}

	c.JSON(http.StatusOK, services)
}

// swagger:operation GET /systems/{system}/build/{serviceId} services GetService
//
// Get service
//
// Get service
// ---
//     consumes:
//     - application/json
//     produces:
//     - application/json
//
//     parameters:
//       - description: System ID
//         in: path
//         name: system
//         required: true
//         type: string
//       - description: Service ID
//         in: path
//         name: serviceId
//         required: true
//         type: string
//
//     responses:
//         '200':
//           description: Service Object
//           schema:
//             "$ref": "#/definitions/Service"
//

// handleGetService handler for GetService
func (api *LatticeAPI) handleGetService(c *gin.Context) {
	systemID := v1.SystemID(c.Param(systemIdentifier))
	serviceID := v1.ServiceID(c.Param(serviceIdentifier))

	service, err := api.backend.Systems().Services(systemID).Get(serviceID)
	if err != nil {
		v1err, ok := err.(*v1.Error)
		if !ok {
			handleInternalError(c, err)
			return
		}

		switch v1err.Code {
		case v1.ErrorCodeInvalidSystemID, v1.ErrorCodeInvalidServiceID:
			c.JSON(http.StatusNotFound, v1err)

		default:
			handleInternalError(c, err)
		}
		return
	}

	c.JSON(http.StatusOK, service)
}

// swagger:operation GET /systems/{system}/services/{serviceId}/logs services GetServiceLogs
//
// Get service logs
//
// Returns a service log stream
// ---
//     consumes:
//     - application/json
//     produces:
//     - application/json
//
//     parameters:
//       - description: System ID
//         in: path
//         name: system
//         required: true
//         type: string
//       - description: Service ID
//         in: path
//         name: serviceId
//         required: true
//         type: string
//       - description: Instance
//         in: query
//         name: instance
//         required: false
//         type: string
//       - description: Sidecar
//         in: query
//         name: sidecar
//         required: false
//         type: string
//       - description: Follow
//         in: query
//         name: follow
//         required: false
//         type: boolean
//       - description: Previous
//         in: query
//         name: previous
//         required: false
//         type: boolean
//       - description: Timestamps
//         in: query
//         name: timestamps
//         required: false
//         type: boolean
//       - description: Tail
//         in: query
//         name: tail
//         required: false
//         type: int
//       - description: Since
//         in: query
//         name: since
//         required: false
//         type: string
//       - description: Since Time
//         in: query
//         name: sinceTime
//         required: false
//         type: string
//     responses:
//         '200':
//           description: log stream
//           schema:
//             type: string

// handleGetServiceLogs handler for GetServiceLogs
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

	log, err := api.backend.Systems().Services(systemID).Logs(serviceId, sidecar, instance, logOptions)
	if err != nil {
		v1err, ok := err.(*v1.Error)
		if !ok {
			handleInternalError(c, err)
			return
		}

		switch v1err.Code {
		case v1.ErrorCodeInvalidSystemID, v1.ErrorCodeInvalidServiceID,
			v1.ErrorCodeInvalidInstance, v1.ErrorCodeInvalidSidecar:
			c.JSON(http.StatusNotFound, v1err)

		default:
			handleInternalError(c, err)
		}
		return
	}

	if log == nil {
		c.Status(http.StatusOK)
		return
	}

	serveLogFile(log, c)
}
