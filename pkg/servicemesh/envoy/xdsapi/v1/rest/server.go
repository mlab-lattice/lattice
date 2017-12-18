package rest

import (
	"fmt"
	"net/http"

	xdsapi "github.com/mlab-lattice/system/pkg/servicemesh/envoy/xdsapi/v1"
	"github.com/mlab-lattice/system/pkg/servicemesh/envoy/xdsapi/v1/cds"
	"github.com/mlab-lattice/system/pkg/servicemesh/envoy/xdsapi/v1/lds"
	"github.com/mlab-lattice/system/pkg/servicemesh/envoy/xdsapi/v1/rds"
	"github.com/mlab-lattice/system/pkg/servicemesh/envoy/xdsapi/v1/sds"

	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
)

type restServer struct {
	router *gin.Engine
	cds    *cds.RequestHandler
	lds    *lds.RequestHandler
	rds    *rds.RequestHandler
	sds    *sds.RequestHandler
}

func RunNewRestServer(b xdsapi.Backend, port int32) {
	s := restServer{
		router: gin.Default(),
		cds:    &cds.RequestHandler{Backend: b},
		lds:    &lds.RequestHandler{Backend: b},
		rds:    &rds.RequestHandler{Backend: b},
		sds:    &sds.RequestHandler{Backend: b},
	}

	s.mountHandlers()

	glog.V(1).Info("Waiting for backend to be ready")
	if !b.Ready() {
		panic("backend Ready() failed")
	}

	s.router.Run(fmt.Sprintf(":%v", port))
}

func (r *restServer) mountHandlers() {
	// CDS
	r.router.GET("/v1/clusters/:service_cluster/:service_node", func(c *gin.Context) {
		serviceCluster := c.Param("service_cluster")
		serviceNode := c.Param("service_node")

		response, err := r.cds.GetResponse(serviceCluster, serviceNode)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}

		c.JSON(http.StatusOK, response)
	})

	// LDS
	r.router.GET("/v1/listeners/:service_cluster/:service_node", func(c *gin.Context) {
		serviceCluster := c.Param("service_cluster")
		serviceNode := c.Param("service_node")

		response, err := r.lds.GetResponse(serviceCluster, serviceNode)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}

		c.JSON(http.StatusOK, response)
	})

	// RDS
	r.router.GET("/v1/routes/:route_config_name/:service_cluster/:service_node", func(c *gin.Context) {
		routeName := c.Param("route_config_name")
		serviceCluster := c.Param("service_cluster")
		serviceNode := c.Param("service_node")

		response, err := r.rds.GetResponse(routeName, serviceCluster, serviceNode)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}

		c.JSON(http.StatusOK, response)
	})

	// SDS
	r.router.GET("/v1/registration/:service_name", func(c *gin.Context) {
		serviceName := c.Param("service_name")

		response, err := r.sds.GetResponse(serviceName)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}

		c.JSON(http.StatusOK, response)
	})
}
