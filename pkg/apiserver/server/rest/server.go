package rest

import (
	"fmt"
	"net/http"

	"github.com/mlab-lattice/system/pkg/apiserver/server"
	"github.com/mlab-lattice/system/pkg/apiserver/server/rest/system"
	"github.com/mlab-lattice/system/pkg/definition/resolver"

	"github.com/gin-gonic/gin"
)

type restServer struct {
	router   *gin.Engine
	backend  server.Backend
	resolver *resolver.SystemResolver
}

func RunNewRestServer(b server.Backend, port int32, workingDirectory string) {
	res, err := resolver.NewSystemResolver(workingDirectory + "/resolver")
	if err != nil {
		panic(err)
	}

	s := restServer{
		router:   gin.Default(),
		backend:  b,
		resolver: res,
	}

	s.mountHandlers()
	s.router.Run(fmt.Sprintf(":%v", port))
}

func (r *restServer) mountHandlers() {
	// Status
	r.router.GET("/status", func(c *gin.Context) {
		c.String(http.StatusOK, "")
	})

	system.MountHandlers(r.router, r.backend, r.resolver)
}
