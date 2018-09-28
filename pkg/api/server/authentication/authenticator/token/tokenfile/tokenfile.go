package tokenfile

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"os"

	"github.com/mlab-lattice/lattice/pkg/api/server/authentication/user"
)

// TokenAuthenticator implementation for authenticator.Token
type TokenAuthenticator struct {
	// tokens map
	tokens map[string]user.User
}

// New creates a new TokenAuthenticator from a token map
func New(tokens map[string]user.User) *TokenAuthenticator {
	return &TokenAuthenticator{
		tokens: tokens,
	}
}

// NewFromCSV creates a new TokenAuthenticator from tokens read from a csv file
func NewFromCSV(path string) (*TokenAuthenticator, error) {
	csvFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	reader := csv.NewReader(bufio.NewReader(csvFile))
	tokens := make(map[string]user.User)

	for {
		line, err := reader.Read()

		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		if len(line) != 2 {
			return nil, fmt.Errorf("bad token values")
		}

		token := line[0]
		name := line[1]
		tokens[token] = user.NewDefaultUser(name)
	}
	return New(tokens), nil
}

// AuthenticateToken
func (a *TokenAuthenticator) AuthenticateToken(token string) (user.User, bool, error) {
	u, exists := a.tokens[token]
	if !exists {
		return nil, false, nil
	}

	return u, true, nil
}
