package test

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	. "github.com/mlab-lattice/lattice/pkg/definition/component/resolver"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	mockresolver "github.com/mlab-lattice/lattice/pkg/backend/mock/definition/component/resolver"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	definitionv1 "github.com/mlab-lattice/lattice/pkg/definition/v1"
	"github.com/mlab-lattice/lattice/pkg/util/git"
	testutil "github.com/mlab-lattice/lattice/pkg/util/test"

	"github.com/mlab-lattice/lattice/pkg/definition/component"
	gitplumbing "gopkg.in/src-d/go-git.v4/plumbing"
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

	service1 = &definitionv1.Service{
		Description: "service 1",
		Container:   container1,
		NodePool:    nodePool1,
	}

	system1 = &definitionv1.System{
		Description: "system 1",
		Components: map[string]component.Interface{
			"job":     job1,
			"service": service1,
		},
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
	type inputComponent struct {
		c     component.Interface
		p     tree.Path
		ctx   *git.CommitReference
		depth int
	}
	type phase struct {
		description string
		repos       []repo
		inputs      map[string]inputComponent
		// maps input names to a map mapping paths to the expected resolution info
		expected map[string]map[tree.Path]*ResolutionInfo
	}
	tests := []struct {
		description string
		phases      []phase
	}{
		{
			description: "no references",
			phases: []phase{
				{
					description: "test",
					inputs: map[string]inputComponent{
						"job": {
							c:     job1,
							p:     tree.RootPath().Child("job"),
							depth: DepthInfinite,
						},
						"service": {
							c:     service1,
							p:     tree.RootPath().Child("service"),
							depth: DepthInfinite,
						},
						"system": {
							c:     system1,
							p:     tree.RootPath(),
							depth: DepthInfinite,
						},
					},
					expected: map[string]map[tree.Path]*ResolutionInfo{
						"job":     {tree.RootPath().Child("job"): {Component: job1}},
						"service": {tree.RootPath().Child("service"): {Component: service1}},
						"system": {
							tree.RootPath():                  {Component: system1},
							tree.RootPath().Child("job"):     {Component: job1},
							tree.RootPath().Child("service"): {Component: service1},
						},
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

		r := NewComponentResolver(gitResolver, mockresolver.NewMemoryTemplateStore(), mockresolver.NewMemorySecretStore())
		for _, phase := range test.phases {
			for _, repo := range phase.repos {
				err := seedRepo(&repo)
				if err != nil {
					t.Fatal(err)
				}
			}

			err := func() error {
				for name, input := range phase.inputs {
					t, err := r.Resolve(input.c, system1ID, input.p, input.ctx, input.depth)
					if err != nil {
						return err
					}

					expected := phase.expected[name]

					if t.Len() != len(expected) {
						return fmt.Errorf("expected %v result to have %v components, but has %v", name, len(expected), t.Len())
					}

					for p, i := range expected {
						info, ok := t.Get(p)
						if !ok {
							return fmt.Errorf("expected %v result to have component at %v but it does not", name, p.String())
						}

						if !reflect.DeepEqual(info, i) {
							return fmt.Errorf(testutil.ErrorDiffsJSON(info, expected))
						}
					}
				}

				return nil
			}()
			if err != nil {
				t.Errorf("test %v, phase %v: %v", test.description, phase.description, err)
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
