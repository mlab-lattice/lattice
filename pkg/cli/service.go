package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/mlab-lattice/system/pkg/types"
)

var serviceHeaders = []string{
	"ID",
	"Path",
	"State",
	"Updated Instances",
	"Stale Instances",
	"Public Ports",
	"Failure Message",
}

func getServiceValues(service *types.Service) []string {
	values := []string{
		string(service.ID),
		string(service.Path),
		string(service.State),
		strconv.Itoa(int(service.UpdatedInstances)),
		strconv.Itoa(int(service.StaleInstances)),
	}

	ports := ""
	for port, info := range service.PublicPorts {
		ports += fmt.Sprintf("%v: %v\n", port, info.Address)
	}

	ports = strings.TrimSpace(ports)
	values = append(values, ports)

	failureMessage := "n/a"
	if service.FailureMessage != nil {
		failureMessage = *service.FailureMessage
	}
	values = append(values, failureMessage)

	return values
}

func ShowService(service *types.Service, output OutputFormat) error {
	switch output {
	case OutputFormatTable:
		values := getServiceValues(service)
		ShowResource(serviceHeaders, values)
	case OutputFormatJSON:
		DisplayAsJSON(service)
	default:
		return NewOutputFormatError(output)
	}
	return nil
}

func ShowServices(services []types.Service, output OutputFormat) error {
	switch output {
	case OutputFormatTable:
		values := make([][]string, len(services))
		for i, b := range services {
			values[i] = getServiceValues(&b)
		}
		ListResources(serviceHeaders, values)
	case OutputFormatJSON:
		DisplayAsJSON(services)
	default:
		return NewOutputFormatError(output)
	}
	return nil
}
