package services

import (
	"fmt"
	"log"

	"github.com/mlab-lattice/system/pkg/cli/latticectl"
	lctlcommand "github.com/mlab-lattice/system/pkg/cli/latticectl/command"
	"github.com/mlab-lattice/system/pkg/managerapi/client"
	"github.com/mlab-lattice/system/pkg/types"
)

type AddressCommand struct {
}

func (c *AddressCommand) Base() (*latticectl.BaseCommand, error) {
	cmd := &lctlcommand.ServiceCommand{
		Name: "address",
		Run: func(ctx lctlcommand.ServiceCommandContext, args []string) {
			GetServiceAddress(ctx.Client().Systems().Services(ctx.SystemID()), ctx.ServiceID())
		},
	}

	return cmd.Base()
}

func GetServiceAddresses(client client.ServiceClient) {
	services, err := client.List()
	if err != nil {
		log.Panic(err)
	}

	for _, service := range services {
		GetServiceAddress(client, service.ID)
	}
}

func GetServiceAddress(client client.ServiceClient, serviceID types.ServiceID) {
	service, err := client.Get(serviceID)
	if err != nil {
		log.Panic(err)
	}

	for port, address := range service.PublicPorts {
		fmt.Printf("%v:%v\n", address.Address, port)
	}
}
