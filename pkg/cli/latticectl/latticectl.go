package latticectl

import (
	"log"

	"github.com/mlab-lattice/system/pkg/managerapi/client"
	"github.com/mlab-lattice/system/pkg/managerapi/client/rest"
)

func DefaultLatticeClient(lattice string) client.Interface {
	return rest.NewClient(lattice)
}

type Latticectl struct {
	Root    Command
	Client  LatticeClientGenerator
	Context ContextManager
}

func (l *Latticectl) Execute() {
	base, err := l.Root.Base()
	if err != nil {
		log.Fatal(err)
	}

	cmd, err := base.Command(l)
	if err != nil {
		log.Fatal(err)
	}

	cmd.Execute()
}

func (l *Latticectl) ExecuteColon() {
	base, err := l.Root.Base()
	if err != nil {
		log.Fatal(err)
	}

	cmd, err := base.Command(l)
	if err != nil {
		log.Fatal(err)
	}

	cmd.ExecuteColon()
}
