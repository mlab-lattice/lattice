package v1

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	v1rest "github.com/mlab-lattice/lattice/pkg/api/v1/rest"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
)

const secretIdentifier = "secret_path"

var secretIdentifierPathComponent = fmt.Sprintf(":%v", secretIdentifier)
var secretPath = fmt.Sprintf(v1rest.SystemSecretPathFormat, systemIdentifierPathComponent, secretIdentifierPathComponent)

var secretsPath = fmt.Sprintf(v1rest.SystemSecretsPathFormat, systemIdentifierPathComponent)

func (api *LatticeAPI) setupSecretsEndpoints() {

	// list-secrets
	api.router.GET(secretsPath, api.handleListSecrets)

	// get-secret
	api.router.GET(secretPath, api.handleGetSecret)

	// set-secret
	api.router.PATCH(secretPath, api.handleSetSecret)

	// unset-secret
	api.router.DELETE(secretPath, api.handleUnsetSecret)

}

// SetSecret godoc
// @ID set-secret
// @Summary set secret
// @Description set secret
// @Router /systems/{system}/secrets [post]
// @Tags secrets
// @Param system path string true "System ID"
// @Param secretRequest body rest.SetSecretRequest true "Create secret"
// @Accept  json
// @Produce  json
// @Success 200 {object} v1.Secret
// @Failure 400 {object} v1.ErrorResponse
func (api *LatticeAPI) handleSetSecret(c *gin.Context) {
	var req v1rest.SetSecretRequest
	if err := c.BindJSON(&req); err != nil {
		handleBadRequestBody(c)
		return
	}

	systemID := v1.SystemID(c.Param(systemIdentifier))
	escapedSecretPath := c.Param(secretIdentifier)

	secretPathString, err := url.PathUnescape(escapedSecretPath)
	if err != nil {
		// FIXME: send invalid secret error
		c.Status(http.StatusBadRequest)
		return
	}

	splitPath := strings.Split(secretPathString, ":")
	if len(splitPath) != 2 {
		// FIXME: send invalid secret error
		c.Status(http.StatusBadRequest)
		return
	}

	path, err := tree.NewNodePath(splitPath[0])
	if err != nil {
		// FIXME: send invalid secret error
		c.Status(http.StatusBadRequest)
		return
	}

	name := splitPath[1]

	err = api.backend.SetSystemSecret(systemID, path, name, req.Value)
	if err != nil {
		handleError(c, err)
		return
	}

	c.Status(http.StatusOK)
}

// ListSecrets godoc
// @ID list-secrets
// @Summary Lists secrets
// @Description list secrets
// @Router /systems/{system}/secrets [get]
// @Tags secrets
// @Param system path string true "System ID"
// @Accept  json
// @Produce  json
// @Success 200 {array} v1.Secret
func (api *LatticeAPI) handleListSecrets(c *gin.Context) {
	systemID := v1.SystemID(c.Param(systemIdentifier))

	secrets, err := api.backend.ListSystemSecrets(systemID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, secrets)
}

// GetSecret godoc
// @ID get-secret
// @Summary Get secret
// @Description get secret
// @Router /systems/{system}/secrets/{secret} [get]
// @Tags secrets
// @Param system path string true "System ID"
// @Param secret path string true "Secret Path"
// @Accept  json
// @Produce  json
// @Success 200 {object} v1.Secret
// @Failure 404 {object} v1.ErrorResponse
func (api *LatticeAPI) handleGetSecret(c *gin.Context) {
	systemID := v1.SystemID(c.Param(systemIdentifier))
	escapedSecretPath := c.Param(secretIdentifier)

	secretPathString, err := url.PathUnescape(escapedSecretPath)
	if err != nil {
		// FIXME: send invalid secret error
		c.Status(http.StatusBadRequest)
		return
	}

	splitPath := strings.Split(secretPathString, ":")
	if len(splitPath) != 2 {
		// FIXME: send invalid secret error
		c.Status(http.StatusBadRequest)
		return
	}

	path, err := tree.NewNodePath(splitPath[0])
	if err != nil {
		// FIXME: send invalid secret error
		c.Status(http.StatusBadRequest)
		return
	}

	name := splitPath[1]

	secret, err := api.backend.GetSystemSecret(systemID, path, name)
	if err != nil {
		// FIXME: send invalid secret error
		c.Status(http.StatusBadRequest)
		return
	}

	c.JSON(http.StatusOK, secret)
}

// DeleteSystem godoc
// @ID delete-system
// @Summary Delete system
// @Description get system
// @Router /systems/{system}/secrets/{secret} [delete]
// @Tags secrets
// @Accept  json
// @Produce  json
// @Param system path string true "System ID"
// @Param secret path string true "Secret Path"
// @Success 200 {object} v1.Result
// @Failure 404 {object} v1.ErrorResponse
func (api *LatticeAPI) handleUnsetSecret(c *gin.Context) {
	systemID := v1.SystemID(c.Param(systemIdentifier))
	escapedSecretPath := c.Param(secretIdentifier)

	secretPathString, err := url.PathUnescape(escapedSecretPath)
	if err != nil {
		// FIXME: send invalid secret error
		c.Status(http.StatusBadRequest)
		return
	}

	splitPath := strings.Split(secretPathString, ":")
	if len(splitPath) != 2 {
		// FIXME: send invalid secret error
		c.Status(http.StatusBadRequest)
		return
	}

	path, err := tree.NewNodePath(splitPath[0])
	if err != nil {
		// FIXME: send invalid secret error
		c.Status(http.StatusBadRequest)
		return
	}

	name := splitPath[1]

	err = api.backend.UnsetSystemSecret(systemID, path, name)
	if err != nil {
		handleError(c, err)
		return
	}

	c.Status(http.StatusOK)

}
