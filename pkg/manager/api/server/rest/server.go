package rest

import (
	"fmt"
	"io"
	"io/ioutil"
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

func logEndpoint(c *gin.Context, log io.ReadCloser, follow bool) {
	defer log.Close()

	if !follow {
		logContents, err := ioutil.ReadAll(log)
		if err != nil {
			c.String(http.StatusInternalServerError, "")
			return
		}
		c.String(http.StatusOK, string(logContents))
		return
	}

	buf := make([]byte, logStreamChunkSize)
	c.Stream(func(w io.Writer) bool {
		n, err := log.Read(buf)
		if err != nil {
			return false
		}

		w.Write(buf[:n])
		return true
	})
}
