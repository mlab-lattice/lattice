package test

import (
	"fmt"
	"os"
	"testing"

	. "github.com/mlab-lattice/lattice/pkg/definition/component/resolver"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	mockresolver "github.com/mlab-lattice/lattice/pkg/backend/mock/definition/component/resolver"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	definitionv1 "github.com/mlab-lattice/lattice/pkg/definition/v1"
	"github.com/mlab-lattice/lattice/pkg/util/git"

	gitplumbing "gopkg.in/src-d/go-git.v4/plumbing"
	"reflect"
)

const workDir = "/tmp/lattice/test/pkg/definition/component/resolver/component_resolver"

var (
	system1ID = v1.SystemID("test")

	container1 = definitionv1.Container{
		Ports: map[int32]definitionv1.ContainerPort{
			80: {
				Protocol: "http",
				ExternalAccess: &definitionv1.ContainerPortExternalAccess{
					Public: true,
				},
			},
		},
		Build: &definitionv1.ContainerBuild{
			CommandBuild: &definitionv1.ContainerBuildCommand{
				BaseImage: definitionv1.DockerImage{
					Repository: "library/ubuntu",
					Tag:        "16.04",
				},
				Command: []string{"npm install", "npm run build"},
			},
		},
	}

	nodePool1Path = tree.PathSubcomponent("/:main")
	nodePool1     = &definitionv1.NodePoolOrReference{
		NodePoolPath: &nodePool1Path,
	}

	job1 = &definitionv1.Job{
		Description: "job 1",
		Container:   container1,
		NodePool:    nodePool1Path,
	}
)

type commit struct {
	contents map[string][]byte
	branch   string
	tag      string
}

type repo struct {
	name    string
	commits map[string]commit
	// maps the name of the commit to its hash
	hashes map[string]gitplumbing.Hash
}

func TestComponentResolver(t *testing.T) {
	type phase struct {
		description string
		repos       []repo
		test        func(ComponentResolver) error
	}
	tests := []struct {
		description string
		phases      []phase
	}{
		{
			description: "no references",
			phases: []phase{
				{
					description: "job",
					test: func(r ComponentResolver) error {
						p := tree.RootPath().Child("job1")
						t, err := r.Resolve(job1, system1ID, p, nil, DepthInfinite)
						if err != nil {
							return err
						}

						if t.Len() != 1 {
							return fmt.Errorf("expected result to have 1 component, but has %v", t.Len())
						}

						i, ok := t.Get(p)
						if !ok {
							return fmt.Errorf("expected result to have component at %v but it does not", p.String())
						}

						j, ok := i.Component.(*definitionv1.Job)
						if !ok {
							return fmt.Errorf("expected component at %v to be a job, it was not", p.String())
						}

						if !reflect.DeepEqual(j, job1) {
							return fmt.Errorf("expected component to match job1, it did not")
						}

						return nil
					},
				},
			},
		},
	}

	for _, test := range tests {
		os.RemoveAll(workDir)
		gitResolver, err := git.NewResolver(fmt.Sprintf("%v/resolver", workDir), true)
		if err != nil {
			t.Fatalf("error creating git resolver: %v", err)
		}

		resolver := NewComponentResolver(gitResolver, mockresolver.NewMemoryTemplateStore(), mockresolver.NewMemorySecretStore())
		for _, phase := range test.phases {
			for _, repo := range phase.repos {
				err := seedRepo(&repo)
				if err != nil {
					t.Fatal(err)
				}
			}

			err := phase.test(resolver)
			if err != nil {
				t.Errorf("test %v phase %v: %v", test.description, phase.description, err)
				break
			}
		}
	}
}

func repoURL(name string) string {
	return fmt.Sprintf("%v/repos/%v", workDir, name)
}

func seedRepo(r *repo) error {
	url := repoURL(r.name)
	hashes := make(map[string]gitplumbing.Hash)
	var lastHash gitplumbing.Hash

	for name, c := range r.commits {
		branch := "master"
		if c.branch != "" {
			branch = c.branch
		}

		if err := git.CheckoutBranch(url, branch); err != nil {
			if err := git.CreateBranch(url, branch, lastHash); err != nil {
				return fmt.Errorf("error initializing repo: %v", err)
			}
		}

		for file, contents := range c.contents {
			err := git.WriteAndAddFile(url, file, contents, 0700)
			if err != nil {
				return fmt.Errorf("error initializing repo: %v", err)
			}
		}
		hash, err := git.Commit(url, name)
		if err != nil {
			return fmt.Errorf("error initializing repo: %v", err)
		}

		hashes[name] = hash
	}

	r.hashes = hashes

	return nil
}
