package latticectl

import (
	"log"

	"github.com/mlab-lattice/system/pkg/apiserver/client"
	"github.com/mlab-lattice/system/pkg/apiserver/client/rest"
	"github.com/mlab-lattice/system/pkg/cli/command"
)

func DefaultLatticeClient(lattice string) client.Interface {
	return rest.NewClient(lattice)
}

type Latticectl struct {
	Root    Command
	Client  ClientFactory
	Context ContextManager
}

func (l *Latticectl) Init() (*command.Command, error) {
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
