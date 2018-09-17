package latticectl

import (
	"log"

	"github.com/mlab-lattice/lattice/pkg/api/client/rest"
	v1client "github.com/mlab-lattice/lattice/pkg/api/client/v1"
	"github.com/mlab-lattice/lattice/pkg/util/cli"
)

func DefaultLatticeClient(lattice string) v1client.Interface {
	// FIXME: add api key support
	return rest.NewUnauthenticatedClient(lattice).V1()
}

type Latticectl struct {
	Root    Command
	Client  ClientFactory
	Context ContextManager
}

func (l *Latticectl) Init() (*cli.Command, error) {
	base, err := l.Root.Base()
	if err != nil {
		return nil, err
	}

	return base.Command(l)
}

func (l *Latticectl) Execute() {
	cmd, err := l.Init()
	if err != nil {
		log.Fatal(err)
	}

	cmd.Execute()
}

func (l *Latticectl) ExecuteColon() {
	cmd, err := l.Init()
	if err != nil {
		log.Fatal(err)
	}

	cmd.ExecuteColon()
}
