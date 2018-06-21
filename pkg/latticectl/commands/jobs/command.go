package jobs

import (
	"bytes"
	"io"
	"log"
	"os"
	"time"

	v1client "github.com/mlab-lattice/lattice/pkg/api/client/v1"
	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/latticectl"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
	"github.com/mlab-lattice/lattice/pkg/util/cli/color"
	"github.com/mlab-lattice/lattice/pkg/util/cli/printer"

	"k8s.io/apimachinery/pkg/util/wait"

	tw "github.com/tfogo/tablewriter"
)

// ListJobsSupportedFormats is the list of printer.Formats supported
// by the ListJobs function.
var ListJobsSupportedFormats = []printer.Format{
	printer.FormatJSON,
	printer.FormatTable,
}

type ListJobsCommand struct {
	Subcommands []latticectl.Command
}

func (c *ListJobsCommand) Base() (*latticectl.BaseCommand, error) {
	output := &latticectl.OutputFlag{
		SupportedFormats: ListJobsSupportedFormats,
	}
	var watch bool
	watchFlag := &latticectl.WatchFlag{
		Target: &watch,
	}

	cmd := &latticectl.SystemCommand{
		Name: "jobs",
		Flags: cli.Flags{
			output.Flag(),
			watchFlag.Flag(),
		},
		Run: func(ctx latticectl.SystemCommandContext, args []string) {
			format, err := output.Value()
			if err != nil {
				log.Fatal(err)
			}

			if watch {
				WatchJobs(ctx.Client().Systems().Jobs(ctx.SystemID()), format, os.Stdout)
			} else {
				err := ListJobs(ctx.Client().Systems().Jobs(ctx.SystemID()), format, os.Stdout)
				if err != nil {
					log.Fatal(err)
				}
			}
		},
		Subcommands: c.Subcommands,
	}

	return cmd.Base()
}

func ListJobs(client v1client.JobClient, format printer.Format, writer io.Writer) error {
	jobs, err := client.List()
	if err != nil {
		return err
	}

	p := jobsPrinter(jobs, format)
	p.Print(writer)
	return nil
}

func WatchJobs(client v1client.JobClient, format printer.Format, writer io.Writer) {
	jobsChan := make(chan []v1.Job)

	lastHeight := 0
	var b bytes.Buffer

	go wait.PollImmediateInfinite(
		5*time.Second,
		func() (bool, error) {
			jobsList, err := client.List()
			if err != nil {
				return false, err
			}

			jobsChan <- jobsList
			return false, nil
		},
	)

	for jobs := range jobsChan {
		p := jobsPrinter(jobs, format)
		lastHeight = p.Overwrite(b, lastHeight)
	}
}

func jobsPrinter(jobs []v1.Job, format printer.Format) printer.Interface {
	var p printer.Interface
	switch format {
	case printer.FormatTable:
		headers := []string{"Path", "State", "Start", "Completion"}

		headerColors := []tw.Colors{
			{tw.Bold},
			{tw.Bold},
			{tw.Bold},
			{tw.Bold},
		}

		columnColors := []tw.Colors{
			{tw.FgHiCyanColor},
			{},
			{},
			{},
		}

		columnAlignment := []int{
			tw.ALIGN_LEFT,
			tw.ALIGN_LEFT,
			tw.ALIGN_LEFT,
			tw.ALIGN_LEFT,
		}

		var rows [][]string
		for _, job := range jobs {
			var stateColor color.Color
			switch job.State {
			case v1.JobStateFailed:
				stateColor = color.Failure
			case v1.JobStateSucceeded:
				stateColor = color.Success
			default:
				stateColor = color.Warning
			}

			startTimestamp := "-"
			if job.StartTimestamp != nil {
				startTimestamp = job.StartTimestamp.String()
			}

			completionTimestamp := "-"
			if job.StartTimestamp != nil {
				completionTimestamp = job.StartTimestamp.String()
			}

			rows = append(rows, []string{
				job.Path.String(),
				stateColor(string(job.State)),
				startTimestamp,
				completionTimestamp,
			})
		}

		p = &printer.Table{
			Headers:         headers,
			Rows:            rows,
			HeaderColors:    headerColors,
			ColumnColors:    columnColors,
			ColumnAlignment: columnAlignment,
		}

	case printer.FormatJSON:
		p = &printer.JSON{
			Value: jobs,
		}
	}

	return p
}
