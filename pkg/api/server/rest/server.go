package rest

import (
	"fmt"
	"net/http"

	restv1 "github.com/mlab-lattice/system/pkg/api/server/rest/v1"
	"github.com/mlab-lattice/system/pkg/api/server/v1"
	"github.com/mlab-lattice/system/pkg/definition/resolver"

	"github.com/gin-gonic/gin"
)

type restServer struct {
	router   *gin.Engine
	backend  v1.Interface
	resolver *resolver.SystemResolver
}

func RunNewRestServer(backend v1.Interface, port int32, workingDirectory string) {
	res, err := resolver.NewSystemResolver(workingDirectory + "/resolver")
	if err != nil {
		panic(err)
	}

	s := restServer{
		router:   gin.Default(),
		backend:  backend,
		resolver: res,
	}

	s.mountHandlers()
	s.router.Run(fmt.Sprintf(":%v", port))
}

func (r *restServer) mountHandlers() {
	// Status
	r.router.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "")
	})

	restv1.MountHandlers(r.router, r.backend, r.resolver)
}
