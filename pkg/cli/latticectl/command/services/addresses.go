package services

import (
	"fmt"
	"io"
	"log"
	"os"
	"sort"

	"github.com/mlab-lattice/system/pkg/cli/command"
	"github.com/mlab-lattice/system/pkg/cli/latticectl"
	lctlcommand "github.com/mlab-lattice/system/pkg/cli/latticectl/command"
	"github.com/mlab-lattice/system/pkg/cli/printer"
	"github.com/mlab-lattice/system/pkg/managerapi/client"
	"github.com/mlab-lattice/system/pkg/types"

	tw "github.com/tfogo/tablewriter"
)

type AddressCommand struct {
}

func (c *AddressCommand) Base() (*latticectl.BaseCommand, error) {
	output := &lctlcommand.OutputFlag{
		SupportedFormats: ListServicesSupportedFormats,
	}

	cmd := &lctlcommand.ServiceCommand{
		Name: "addresses",
		Flags: command.Flags{
			output.Flag(),
		},
		Run: func(ctx lctlcommand.ServiceCommandContext, args []string) {
			format, err := output.Value()
			if err != nil {
				log.Fatal(err)
			}

			err = GetServiceAddress(ctx.Client().Systems().Services(ctx.SystemID()), ctx.ServiceID(), format, os.Stdout)
			if err != nil {
				log.Fatal(err)
			}
		},
	}

	return cmd.Base()
}

func GetServiceAddresses(client client.ServiceClient, format printer.Format, writer io.Writer) error {
	services, err := client.List()
	if err != nil {
		return err
	}

	for _, service := range services {
		servicePath := service.Path.ToDomain(true)

		err = GetServiceAddress(client, types.ServiceID(servicePath), format, writer)
		if err != nil {
			return err
		}
	}

	return nil
}

func GetServiceAddress(client client.ServiceClient, serviceID types.ServiceID, format printer.Format, writer io.Writer) error {
	service, err := client.Get(serviceID)
	if err != nil {
		return err
	}

	p := AddressPrinter(service, format)
	p.Print(writer)

	return nil
}

func AddressPrinter(service *types.Service, format printer.Format) printer.Interface {
	var p printer.Interface
	switch format {
	case printer.FormatDefault, printer.FormatTable:
		headers := []string{"Port", "Address"}

		headerColors := []tw.Colors{
			{tw.Bold},
			{tw.Bold},
		}

		columnColors := []tw.Colors{
			{},
			{},
		}

		columnAlignment := []int{
			tw.ALIGN_LEFT,
			tw.ALIGN_LEFT,
		}

		var rows [][]string
		for port, address := range service.PublicPorts {
			rows = append(rows, []string{
				fmt.Sprintf("%d", port),
				string(address.Address),
			})
		}

		sort.Slice(rows, func(i, j int) bool { return rows[i][0] < rows[j][0] })

		p = &printer.Table{
			Headers:         headers,
			Rows:            rows,
			HeaderColors:    headerColors,
			ColumnColors:    columnColors,
			ColumnAlignment: columnAlignment,
		}

	case printer.FormatJSON:
		p = &printer.JSON{
			Value: service.PublicPorts,
		}
	}

	return p
}
