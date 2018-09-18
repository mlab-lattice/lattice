package rest

import (
	"fmt"
	"net/http"

	"github.com/mlab-lattice/lattice/pkg/api/server/auth"
	restv1 "github.com/mlab-lattice/lattice/pkg/api/server/rest/v1"
	"github.com/mlab-lattice/lattice/pkg/api/server/v1"
	"github.com/mlab-lattice/lattice/pkg/api/users"
	"github.com/mlab-lattice/lattice/pkg/definition/resolver"

	"github.com/gin-gonic/gin"
)

const (
	currentUserContextKey = "CURRENT_USER"
)

type restServer struct {
	router         *gin.Engine
	backend        v1.Interface
	resolver       resolver.ComponentResolver
	authenticators []auth.Authenticator
}

func RunNewRestServer(backend v1.Interface, resolver resolver.ComponentResolver, port int32, options *ServerOptions) {
	router := gin.Default()
	// Some of our paths use URL encoded paths, so don't have
	// gin decode those
	router.UseRawPath = true
	s := restServer{
		router:   router,
		backend:  backend,
		resolver: resolver,
	}
	s.initAuthenticators(options)

	s.mountHandlers(options)
	s.router.Run(fmt.Sprintf(":%v", port))
}
func (r *restServer) initAuthenticators(options *ServerOptions) {
	authenticators := make([]auth.Authenticator, 0)

	// setup legacy authentication as needed
	if options.AuthOptions.AuthType == AuthTypeLegacy {
		fmt.Println("Setting up authentication with legacy api key header")
		authenticators = append(authenticators, auth.NewLegacyApiKeyAuthenticator(options.AuthOptions.LegacyApiAuthKey))
	}
	r.authenticators = authenticators
}
func (r *restServer) mountHandlers(options *ServerOptions) {
	// Status
	r.router.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "")
	})

	routerGroup := r.router.Group("/")
	// setup api key authentication if specified

	restv1.MountHandlers(routerGroup, r.backend, r.resolver)
}

func (r *restServer) setupAuthentication(router *gin.RouterGroup) {
	if len(r.authenticators) == 0 {
		fmt.Println("WARNING: No authenticators configured.")
	} else {
		router.Use(r.authenticateRequest())
	}

}

// authenticateRequest authenticates the request against the configured authentication api key
func (r *restServer) authenticateRequest() gin.HandlerFunc {
	return func(c *gin.Context) {
		for _, authenticator := range r.authenticators {
			userObject, err := authenticator.AuthenticateRequest(c)

			if err != nil {
				c.AbortWithStatusJSON(http.StatusForbidden,
					gin.H{"error": fmt.Sprintf("Failed to authenticate: %v", err)})
				return
			}
			if userObject != nil { // Auth Success!
				fmt.Printf("User %v successfully authenticated\n", userObject.GetName())
				// Attach user to current context
				c.Set(currentUserContextKey, userObject)
				return
			}

		}

		// Authentication failure
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "No authentication provided"})
	}
}

func GetCurrentUser(c *gin.Context) *users.User {
	if currentUser, exists := c.Get(currentUserContextKey); exists {
		return currentUser.(*users.User)
	}
	return nil
}
