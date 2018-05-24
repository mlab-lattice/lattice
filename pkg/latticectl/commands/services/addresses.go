package services

import (
	"fmt"
	"io"
	"log"
	"os"
	"sort"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/latticectl"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
	"github.com/mlab-lattice/lattice/pkg/util/cli/printer"

	tw "github.com/tfogo/tablewriter"
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

			service, err := lookupService(ctx)
			err = PrintServiceAddress(service, format, os.Stdout)
			if err != nil {
				log.Fatal(err)
			}
		},
	}

	return cmd.Base()
}

func PrintServiceAddress(service *v1.Service, format printer.Format, writer io.Writer) error {

	p := AddressPrinter(service, format)
	p.Print(writer)

	return nil
}

func AddressPrinter(service *v1.Service, format printer.Format) printer.Interface {
	var p printer.Interface
	switch format {
	case printer.FormatTable:
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
		for port, address := range service.Ports {
			rows = append(rows, []string{
				fmt.Sprintf("%d", port),
				address,
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
			Value: service.Ports,
		}
	}

	return p
}
