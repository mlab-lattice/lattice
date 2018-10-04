package oidc

import (
	"fmt"

	"github.com/coreos/go-oidc"
	"github.com/mlab-lattice/lattice/pkg/api/server/authentication/user"
	"golang.org/x/net/context"
)

// Authenticator a token authenticator (implements pkg/api/server/authentication/authenticator.Token interface) that authenticates using Open ID Connect (oidc)
type Authenticator struct {
	issuerURL     string
	clientID      string
	usernameClaim string
}

func New(issuerURL string, clientID string, usernameClaim string) *Authenticator {
	return &Authenticator{
		issuerURL:     issuerURL,
		clientID:      clientID,
		usernameClaim: usernameClaim,
	}
}

func (a *Authenticator) AuthenticateToken(token string) (user.User, bool, error) {

	ctx := context.Background()
	provider, err := oidc.NewProvider(ctx, a.issuerURL)
	if err != nil {
		return nil, false, err
	}
	oidcConfig := &oidc.Config{
		ClientID: a.clientID,
	}
	verifier := provider.Verifier(oidcConfig)

	idToken, err := verifier.Verify(ctx, token)

	if err != nil {
		return nil, false, err
	}

	var c claims
	if err := idToken.Claims(&c); err != nil {
		return nil, false, fmt.Errorf("oidc: parse claims: %v", err)
	}

	// validate email
	if a.usernameClaim == "email" {
		// validate username
		if c.Email == "" {
			return nil, false, fmt.Errorf("oidc: no email in id token")
		}

		if !c.EmailVerified {
			return nil, false, fmt.Errorf("oidc: email not verified")
		}
	} else {
		return nil, false, fmt.Errorf("oidc: usernameClaim '%s' not support yet", a.usernameClaim)
	}

	fmt.Printf("oidc: successfully authenticated user '%v'\n", c.Email)
	// return user
	return user.NewDefaultUser(c.Email), true, nil
}

type claims struct {
	Name          string `json:"name"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
}
