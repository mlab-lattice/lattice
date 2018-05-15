package services

import (
	"fmt"
	"io"
	"log"
	"os"
	"sort"

	v1client "github.com/mlab-lattice/lattice/pkg/api/client/v1"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	"github.com/mlab-lattice/lattice/pkg/latticectl"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
	"github.com/mlab-lattice/lattice/pkg/util/cli/printer"
)

type AddressCommand struct {
}

func (c *AddressCommand) Base() (*latticectl.BaseCommand, error) {
	output := &latticectl.OutputFlag{
		SupportedFormats: ListServicesSupportedFormats,
	}

	cmd := &latticectl.ServiceCommand{
		Name: "addresses",
		Flags: cli.Flags{
			output.Flag(),
		},
		Run: func(ctx latticectl.ServiceCommandContext, args []string) {
			format, err := output.Value()
			if err != nil {
				log.Fatal(err)
			}

			err = GetServiceAddress(ctx.Client().Systems().Services(ctx.SystemID()), ctx.ServicePath(), format, os.Stdout)
			if err != nil {
				log.Fatal(err)
			}
		},
	}

	return cmd.Base()
}

func GetServiceAddresses(client v1client.ServiceClient, format printer.Format, writer io.Writer) error {
	services, err := client.List()
	if err != nil {
		return err
	}

	for _, service := range services {
		err = GetServiceAddress(client, service.Path, format, writer)
		if err != nil {
			return err
		}
	}

	return nil
}

func GetServiceAddress(client v1client.ServiceClient, servicePath tree.NodePath, format printer.Format, writer io.Writer) error {
	service, err := client.Get(servicePath)
	if err != nil {
		return err
	}

	p := AddressPrinter(service, format)
	p.Print(writer)

	return nil
}

func AddressPrinter(service *v1.Service, format printer.Format) printer.Interface {
	var p printer.Interface
	switch format {
	case printer.FormatTable:
		var rows [][]string
		headers := []string{"Port", "Address"}

		for port, address := range service.Ports {
			rows = append(rows, []string{
				fmt.Sprintf("%d", port),
				address,
			})
		}

		sort.Slice(rows, func(i, j int) bool { return rows[i][0] < rows[j][0] })

		p = &printer.Table{
			Headers: headers,
			Rows:    rows,
		}

	case printer.FormatJSON:
		p = &printer.JSON{
			Value: service.Ports,
		}
	}

	return p
}
