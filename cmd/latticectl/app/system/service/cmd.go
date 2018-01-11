package service

import (
	"fmt"
	"log"
	"os"

	"github.com/mlab-lattice/system/pkg/cli"
	"github.com/mlab-lattice/system/pkg/constants"
	"github.com/mlab-lattice/system/pkg/definition/tree"
	"github.com/mlab-lattice/system/pkg/managerapi/client"
	"github.com/mlab-lattice/system/pkg/managerapi/client/rest"
	"github.com/mlab-lattice/system/pkg/types"

	"github.com/spf13/cobra"
	"strconv"
	"strings"
)

var (
	output string

	systemIDString string
	url            string
	systemID       types.SystemID
	userClient     client.Interface
	systemClient   client.SystemClient
)

var Cmd = &cobra.Command{
	Use:  "service",
	Args: cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
		os.Exit(1)
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "list services",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		services, err := systemClient.Services(systemID).List()
		if err != nil {
			log.Panic(err)
		}

		if err := cli.ShowServices(services, cli.OutputFormat(output)); err != nil {
			log.Panic(err)
		}
	},
}

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "get service",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		pathString := args[0]
		path, err := tree.NewNodePath(pathString)
		if err != nil {
			log.Panic(err)
		}

		service, err := systemClient.Services(systemID).Get(types.ServiceID(path.ToDomain(true)))
		if err != nil {
			log.Panic(err)
		}

		if err := cli.ShowService(service, cli.OutputFormat(output)); err != nil {
			log.Panic(err)
		}
	},
}

var addressCmd = &cobra.Command{
	Use:   "address",
	Short: "address [/service/path:port]",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		pathAddress := args[0]
		pathAddressParts := strings.Split(pathAddress, ":")
		if len(pathAddressParts) != 2 {
			log.Fatal("invalid string address format")
		}

		pathString := pathAddressParts[0]

		path, err := tree.NewNodePath(pathString)
		if err != nil {
			log.Fatal(err)
		}

		service, err := systemClient.Services(systemID).Get(types.ServiceID(path.ToDomain(true)))
		if err != nil {
			log.Fatal(err)
		}

		portString := pathAddressParts[1]
		port, err := strconv.Atoi(portString)
		if err != nil {
			log.Fatal(err)
		}

		info, ok := service.PublicPorts[int32(port)]
		if !ok {
			log.Fatal("invalid port: " + string(port))
		}

		fmt.Println(info.Address)
	},
}

func init() {
	cobra.OnInitialize(initCmd)

	Cmd.PersistentFlags().StringVar(&url, "url", "", "URL of the manager-api for the system")
	Cmd.PersistentFlags().StringVar(&systemIDString, "system", string(constants.SystemIDDefault), "system to use")

	Cmd.AddCommand(listCmd)
	listCmd.Flags().StringVarP(&output, "output", "o", "table", "whether or not to display output as JSON")
	Cmd.AddCommand(getCmd)
	getCmd.Flags().StringVarP(&output, "output", "o", "table", "whether or not to display output as JSON")
	Cmd.AddCommand(addressCmd)
}

func initCmd() {
	systemID = types.SystemID(systemIDString)

	userClient = rest.NewClient(url)
	systemClient = userClient.Systems()
}
