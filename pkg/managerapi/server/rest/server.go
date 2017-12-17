package rest

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/mlab-lattice/system/pkg/definition/resolver"
	"github.com/mlab-lattice/system/pkg/managerapi/server/user"

	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
)

const (
	// FIXME: this was totally arbitrary. figure out a better size
	logStreamChunkSize = 1024 * 4
)

type restServer struct {
	router   *gin.Engine
	backend  user.Backend
	resolver *resolver.SystemResolver
}

func RunNewRestServer(b user.Backend, port int32, workingDirectory string) {
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

	r.mountSystemHandlers()
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

func handleInternalError(c *gin.Context, err error) {
	glog.Errorf("encountered error: %v", err.Error())
	c.String(http.StatusInternalServerError, "")
}
