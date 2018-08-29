package app

import (
	goflag "flag"

	"github.com/mlab-lattice/lattice/pkg/api/server/mock"
	"github.com/mlab-lattice/lattice/pkg/util/cli"

	"github.com/spf13/pflag"
)

func Command() *cli.Command {
	// https://flowerinthenight.com/blog/2017/12/01/golang-cobra-glog
	pflag.CommandLine.AddGoFlagSet(goflag.CommandLine)

	var port int32
	var apiAuthKey string

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
		},
		Run: func(args []string) {

			mock.RunMockNewRestServer(port, apiAuthKey)
		},
	}

	return command
}
