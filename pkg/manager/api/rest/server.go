package rest

import (
	"fmt"
	"net/http"

	systemresolver "github.com/mlab-lattice/core/pkg/system/resolver"

	"github.com/mlab-lattice/system/pkg/manager/backend"

	"github.com/gin-gonic/gin"
)

const (
	// FIXME: this was totally arbitrary. figure out a better size
	logStreamChunkSize = 1024 * 4
)

type restServer struct {
	router   *gin.Engine
	backend  backend.Interface
	resolver *systemresolver.SystemResolver
}

func RunNewRestServer(b backend.Interface, port int32, workingDirectory string) {
	resolver, err := systemresolver.NewSystemResolver(workingDirectory + "/resolver")
	if err != nil {
		panic(err)
	}

	s := restServer{
		router:   gin.Default(),
		backend:  b,
		resolver: resolver,
	}

	s.mountHandlers()
	s.router.Run(fmt.Sprintf(":%v", port))
}

func (r *restServer) mountHandlers() {
	// Status
	r.router.GET("/status", func(c *gin.Context) {
		c.String(http.StatusOK, "")
	})

	r.mountNamespaceHandlers()
	r.mountAdminHandlers()
}
