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
			err := GetServiceAddress(ctx.Client().Systems().Services(ctx.SystemID()), ctx.ServiceID())
			if err != nil {
				log.Fatal(err)
			}
		},
	}

	return cmd.Base()
}

func GetServiceAddresses(client client.ServiceClient) error {
	services, err := client.List()
	if err != nil {
		return err
	}

	for _, service := range services {
		servicePath := service.Path.ToDomain(true)

		err = GetServiceAddress(client, types.ServiceID(servicePath))
		if err != nil {
			return err
		}
	}

	return nil
}

func GetServiceAddress(client client.ServiceClient, serviceID types.ServiceID) error {
	service, err := client.Get(serviceID)
	if err != nil {
		return err
	}

	for port, address := range service.PublicPorts {
		fmt.Printf("%v:%v\n", address.Address, port)
	}
	return nil
}
