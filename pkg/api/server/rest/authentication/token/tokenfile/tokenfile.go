package tokenfile

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"os"

	"github.com/mlab-lattice/lattice/pkg/api/server/rest/authentication/user"
)

type TokenAuthenticator struct {
	tokens map[string]user.User
}

func New(tokens map[string]user.User) *TokenAuthenticator {
	return &TokenAuthenticator{
		tokens: tokens,
	}
}

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
		username := line[1]

		tokens[token] = &user.DefaultUser{
			Username: username,
		}
	}
	return New(tokens), nil
}

func (authenticator *TokenAuthenticator) AuthenticateToken(token string) (user.User, bool, error) {
	u, exists := authenticator.tokens[token]
	if !exists {
		return nil, false, fmt.Errorf("no such token")
	}

	return u, true, nil
}
