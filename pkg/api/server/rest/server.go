package rest

import (
	"fmt"
	"net/http"

	restv1 "github.com/mlab-lattice/lattice/pkg/api/server/rest/v1"
	"github.com/mlab-lattice/lattice/pkg/api/server/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/resolver"

	"github.com/gin-gonic/gin"
)

const (
	apiKeyHeader = "API_KEY"
)

type restServer struct {
	router   *gin.Engine
	backend  v1.Interface
	resolver *resolver.SystemResolver
}

func RunNewRestServer(backend v1.Interface, port int32, workingDirectory string, apiAuthKey string) {
	res, err := resolver.NewSystemResolver(workingDirectory + "/resolver")
	if err != nil {
		panic(err)
	}

	fmt.Sprintf("server.go: API AUTH KEY '%s'\n", apiAuthKey)

	router := gin.Default()
	// Some of our paths use URL encoded paths, so don't have
	// gin decode those
	router.UseRawPath = true
	s := restServer{
		router:   router,
		backend:  backend,
		resolver: res,
	}

	s.mountHandlers(apiAuthKey)
	s.router.Run(fmt.Sprintf(":%v", port))
}

func (r *restServer) mountHandlers(apiAuthKey string) {
	// Status
	r.router.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "")
	})

	routerGroup := r.router.Group("/")
	fmt.Printf("apiAuthKey: '%s'\n", apiAuthKey)
	// setup api key authentication if specified
	if apiAuthKey != "" {
		fmt.Printf("Setting up authentication with api key header %s\n", apiKeyHeader)
		routerGroup.Use(authenticateRequest(apiAuthKey))
	} else {
		fmt.Println("WARNING: Api key authentication not set")
	}

	restv1.MountHandlers(routerGroup, r.backend, r.resolver)
}

// authenticateRequest authenticates the request against the configured authentication api key
func authenticateRequest(apiAuthKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Println(c.Request.URL.Path)
		// grab request API key from header
		requestApiKey := c.Request.Header.Get(apiKeyHeader)
		if requestApiKey == "" {
			fmt.Printf("WARNING: Auth failure: %s header is not set\n", apiKeyHeader)
		} else if requestApiKey != apiAuthKey {
			fmt.Printf("WARNING: Auth failure: invalid %s\n", apiKeyHeader)
			// TODO enable when all clients provide authentication key
			//c.JSON(http.StatusForbidden, gin.H{"error": "Invalid API_KEY"})
		} else {
			fmt.Println("Auth SUCCESS!")
		}
	}
}
