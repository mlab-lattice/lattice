package test

import (
	"fmt"
	//"os"
	"reflect"
	"testing"

	. "github.com/mlab-lattice/lattice/pkg/definition/component/resolver"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	mockresolver "github.com/mlab-lattice/lattice/pkg/backend/mock/definition/component/resolver"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	definitionv1 "github.com/mlab-lattice/lattice/pkg/definition/v1"
	"github.com/mlab-lattice/lattice/pkg/util/git"
	testutil "github.com/mlab-lattice/lattice/pkg/util/test"

	"encoding/json"
	"github.com/mlab-lattice/lattice/pkg/definition/component"
	//gitplumbing "gopkg.in/src-d/go-git.v4/plumbing"
	"os"
)

const workDir = "/tmp/lattice/test/pkg/definition/component/resolver/component_resolver"

var (
	system1ID = v1.SystemID("test")
	repo1     = "repo1"

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

	container2 = definitionv1.Container{
		Ports: map[int32]definitionv1.ContainerPort{
			80: {
				Protocol: "tcp",
			},
		},
		Build: &definitionv1.ContainerBuild{
			CommandBuild: &definitionv1.ContainerBuildCommand{
				BaseImage: definitionv1.DockerImage{
					Repository: "library/ubuntu",
					Tag:        "16.04",
				},
				Command: []string{"./install.sh"},
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
	job1Bytes, _ = json.Marshal(&job1)

	service1 = &definitionv1.Service{
		Description: "service 1",
		Container:   container1,
		NodePool:    nodePool1,
	}
	service1Bytes, _ = json.Marshal(&service1)

	service2 = &definitionv1.Service{
		Description: "service 2",
		Container:   container2,
		NodePool:    nodePool1,
	}
	service2Bytes, _ = json.Marshal(&service1)

	system1 = &definitionv1.System{
		Description: "system 1",
		Components: map[string]component.Interface{
			"job":     job1,
			"service": service1,
		},
	}
	system1Bytes, _ = json.Marshal(&system1)
)

func TestComponentResolver_NoReferences(t *testing.T) {
	r := resolver()

	// job
	result, err := r.Resolve(job1, system1ID, tree.RootPath(), nil, DepthInfinite)
	if err != nil {
		t.Errorf("expected no error resolving plain job, got :%v", err)
	}

	expected := NewComponentTree()
	expected.Insert(tree.RootPath(), &ResolutionInfo{Component: job1})
	compareComponentTrees(t, "job", expected, result)

	// service
	result, err = r.Resolve(service1, system1ID, tree.RootPath(), nil, DepthInfinite)
	if err != nil {
		t.Errorf("expected no error resolving plain job, got :%v", err)
	}

	expected = NewComponentTree()
	expected.Insert(tree.RootPath(), &ResolutionInfo{Component: service1})
	compareComponentTrees(t, "service", expected, result)

	// system
	result, err = r.Resolve(system1, system1ID, tree.RootPath(), nil, DepthInfinite)
	if err != nil {
		t.Errorf("expected no error resolving plain job, got :%v", err)
	}

	expected = NewComponentTree()
	expected.Insert(tree.RootPath(), &ResolutionInfo{Component: system1})
	expected.Insert(tree.RootPath().Child("job"), &ResolutionInfo{Component: job1})
	expected.Insert(tree.RootPath().Child("service"), &ResolutionInfo{Component: service1})
	compareComponentTrees(t, "system", expected, result)
}

func TestComponentResolver_CommitReference(t *testing.T) {
	r := resolver()

	// seed repo
	repo := repoURL(repo1)
	os.RemoveAll(repo)
	err := git.Init(repo)
	if err != nil {
		t.Fatalf("error initializing repo: %v", err)
	}

	jobCommit, err := git.WriteAndCommitFile(
		repo,
		DefaultFile,
		job1Bytes,
		0700,
		"job",
	)
	if err != nil {
		t.Fatalf("error commiting to repo: %v", err)
	}

	serviceCommit, err := git.WriteAndCommitFile(
		repo,
		DefaultFile,
		service1Bytes,
		0700,
		"job",
	)
	if err != nil {
		t.Fatalf("error commiting to repo: %v", err)
	}

	systemCommit, err := git.WriteAndCommitFile(
		repo,
		DefaultFile,
		system1Bytes,
		0700,
		"job",
	)
	if err != nil {
		t.Fatalf("error commiting to repo: %v", err)
	}

	// job
	jobCommitStr := jobCommit.String()
	ctx := &git.CommitReference{
		RepositoryURL: repo,
		Commit:        jobCommitStr,
	}
	ref := &definitionv1.Reference{
		GitRepository: &definitionv1.GitRepositoryReference{
			GitRepository: &definitionv1.GitRepository{
				URL:    repo,
				Commit: &jobCommitStr,
			},
		},
	}
	result, err := r.Resolve(ref, system1ID, tree.RootPath(), nil, DepthInfinite)
	if err != nil {
		t.Errorf("expected no error resolving plain job, got :%v", err)
	}
	expected := NewComponentTree()
	expected.Insert(tree.RootPath(), &ResolutionInfo{
		Component: job1,
		Commit:    ctx,
	})
	compareComponentTrees(t, "job", expected, result)

	// service
	serviceCommitStr := serviceCommit.String()
	ctx = &git.CommitReference{
		RepositoryURL: repo,
		Commit:        serviceCommitStr,
	}
	ref = &definitionv1.Reference{
		GitRepository: &definitionv1.GitRepositoryReference{
			GitRepository: &definitionv1.GitRepository{
				URL:    repo,
				Commit: &serviceCommitStr,
			},
		},
	}
	result, err = r.Resolve(ref, system1ID, tree.RootPath(), nil, DepthInfinite)
	if err != nil {
		t.Errorf("expected no error resolving plain job, got :%v", err)
	}
	expected = NewComponentTree()
	expected.Insert(tree.RootPath(), &ResolutionInfo{
		Component: service1,
		Commit:    ctx,
	})
	compareComponentTrees(t, "service", expected, result)

	//system
	systemCommitStr := systemCommit.String()
	ctx = &git.CommitReference{
		RepositoryURL: repo,
		Commit:        systemCommitStr,
	}
	ref = &definitionv1.Reference{
		GitRepository: &definitionv1.GitRepositoryReference{
			GitRepository: &definitionv1.GitRepository{
				URL:    repo,
				Commit: &systemCommitStr,
			},
		},
	}
	result, err = r.Resolve(ref, system1ID, tree.RootPath(), nil, DepthInfinite)
	if err != nil {
		t.Errorf("expected no error resolving plain job, got :%v", err)
	}
	expected = NewComponentTree()
	expected.Insert(tree.RootPath(), &ResolutionInfo{
		Component: system1,
		Commit:    ctx,
	})
	expected.Insert(tree.RootPath().Child("job"), &ResolutionInfo{
		Component: job1,
		Commit:    ctx,
	})
	expected.Insert(tree.RootPath().Child("service"), &ResolutionInfo{
		Component: service1,
		Commit:    ctx,
	})
	compareComponentTrees(t, "service", expected, result)

	// invalid commit
	invalidCommit := "0123456789abcdef0123456789abcdef01234567"
	ref = &definitionv1.Reference{
		GitRepository: &definitionv1.GitRepositoryReference{
			GitRepository: &definitionv1.GitRepository{
				URL:    repo,
				Commit: &invalidCommit,
			},
		},
	}
	_, err = r.Resolve(ref, system1ID, tree.RootPath(), nil, DepthInfinite)
	if err == nil {
		t.Errorf("expected error resolving invalid commit but got none")
	}
}

func TestComponentResolver_BranchReference(t *testing.T) {
	r := resolver()

	// seed repo
	repo := repoURL(repo1)
	os.RemoveAll(repo)
	err := git.Init(repo)
	if err != nil {
		t.Fatalf("error initializing repo: %v", err)
	}

	jobCommit, err := git.WriteAndCommitFile(
		repo,
		DefaultFile,
		job1Bytes,
		0700,
		"job",
	)
	if err != nil {
		t.Fatalf("error commiting to repo: %v", err)
	}

	devBranch := "dev"
	err = git.CreateBranch(repo, devBranch, jobCommit)
	if err != nil {
		t.Fatalf("error checking out branch: %v", err)
	}

	serviceCommit, err := git.WriteAndCommitFile(
		repo,
		DefaultFile,
		service1Bytes,
		0700,
		"job",
	)
	if err != nil {
		t.Fatalf("error commiting to repo: %v", err)
	}

	ctx := &git.CommitReference{
		RepositoryURL: repo,
		Commit:        serviceCommit.String(),
	}
	ref := &definitionv1.Reference{
		GitRepository: &definitionv1.GitRepositoryReference{
			GitRepository: &definitionv1.GitRepository{
				URL:    repo,
				Branch: &devBranch,
			},
		},
	}
	result, err := r.Resolve(ref, system1ID, tree.RootPath(), nil, DepthInfinite)
	if err != nil {
		t.Errorf("expected no error resolving branch, got :%v", err)
	}
	expected := NewComponentTree()
	expected.Insert(tree.RootPath(), &ResolutionInfo{
		Component: service1,
		Commit:    ctx,
	})
	compareComponentTrees(t, "service", expected, result)

	// commit again and see the reference be updated
	systemCommit, err := git.WriteAndCommitFile(
		repo,
		DefaultFile,
		system1Bytes,
		0700,
		"job",
	)
	if err != nil {
		t.Fatalf("error commiting to repo: %v", err)
	}

	ctx = &git.CommitReference{
		RepositoryURL: repo,
		Commit:        systemCommit.String(),
	}
	result, err = r.Resolve(ref, system1ID, tree.RootPath(), nil, DepthInfinite)
	if err != nil {
		t.Errorf("expected no error resolving branch, got :%v", err)
	}

	expected = NewComponentTree()
	expected.Insert(tree.RootPath(), &ResolutionInfo{
		Component: system1,
		Commit:    ctx,
	})
	expected.Insert(tree.RootPath().Child("job"), &ResolutionInfo{
		Component: job1,
		Commit:    ctx,
	})
	expected.Insert(tree.RootPath().Child("service"), &ResolutionInfo{
		Component: service1,
		Commit:    ctx,
	})
	compareComponentTrees(t, "system", expected, result)

	// invalid branch
	invalidBranch := "foo"
	ref = &definitionv1.Reference{
		GitRepository: &definitionv1.GitRepositoryReference{
			GitRepository: &definitionv1.GitRepository{
				URL:    repo,
				Branch: &invalidBranch,
			},
		},
	}
	_, err = r.Resolve(ref, system1ID, tree.RootPath(), nil, DepthInfinite)
	if err == nil {
		t.Errorf("expected error resolving invalid commit but got none")
	}
}

func compareComponentTrees(t *testing.T, name string, expected, actual *ComponentTree) {
	if expected.Len() != actual.Len() {
		t.Errorf("expected %v result to contain %v entries, found %v", name, expected.Len(), actual.Len())
		return
	}

	expected.Walk(func(path tree.Path, info *ResolutionInfo) tree.WalkContinuation {
		result, ok := actual.Get(path)
		if !ok {
			t.Errorf("expected %v result to contain %v but it did not", name, path.String())
			return tree.ContinueWalk
		}

		if !reflect.DeepEqual(info, result) {
			t.Errorf("result for %v path %v did not match expected \n%v", name, path.String(), testutil.ErrorDiffsJSON(info, result))
		}

		return tree.ContinueWalk
	})
}

func repoURL(name string) string {
	return fmt.Sprintf("%v/repos/%v", workDir, name)
}

func resolver() ComponentResolver {
	gitResolver, err := git.NewResolver(fmt.Sprintf("%v/resolver", workDir), true)
	if err != nil {
		panic(err)
	}

	return NewComponentResolver(gitResolver, mockresolver.NewMemoryTemplateStore(), mockresolver.NewMemorySecretStore())
}
