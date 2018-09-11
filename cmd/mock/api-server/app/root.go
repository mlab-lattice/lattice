package app

import (
	goflag "flag"

	"github.com/mlab-lattice/lattice/pkg/api/server/rest"
	mockbackend "github.com/mlab-lattice/lattice/pkg/backend/mock/api/server/backend"
	mockresolver "github.com/mlab-lattice/lattice/pkg/backend/mock/definition/component/resolver"
	"github.com/mlab-lattice/lattice/pkg/definition/component/resolver"
	"github.com/mlab-lattice/lattice/pkg/util/cli"

	"github.com/mlab-lattice/lattice/pkg/util/git"
	"github.com/spf13/pflag"
)

func Command() *cli.Command {
	// https://flowerinthenight.com/blog/2017/12/01/golang-cobra-glog
	pflag.CommandLine.AddGoFlagSet(goflag.CommandLine)

	var port int32
	var apiAuthKey string
	var workDirectory string

	command := &cli.Command{
		Name: "api-server",
		Flags: cli.Flags{
			&cli.Int32Flag{
				Name:    "port",
				Usage:   "port to bind to",
				Default: 8080,
				Target:  &port,
			},
			&cli.StringFlag{
				Name:    "api-auth-key",
				Usage:   "if supplied, the required value of the API_KEY header",
				Default: "",
				Target:  &apiAuthKey,
			},
			&cli.StringFlag{
				Name:    "work-directory",
				Usage:   "directory used to download git repositories",
				Default: "/tmp/lattice/mock/api-server",
				Target:  &workDirectory,
			},
		},
		Run: func(args []string) {
			backend := mockbackend.NewMockBackend()

			templateStore := mockresolver.NewMemoryTemplateStore()
			secretStore := mockresolver.NewMemorySecretStore()
			gitResolver, err := git.NewResolver(workDirectory, false)
			if err != nil {
				panic(err)
			}

			r := resolver.NewComponentResolver(gitResolver, templateStore, secretStore)
			rest.RunNewRestServer(backend, r, port, apiAuthKey)
		},
	}

	return command
}
