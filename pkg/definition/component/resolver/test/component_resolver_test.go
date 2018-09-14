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

	job2 = &definitionv1.Job{
		Description: "job 2",
		Container:   container2,
		NodePool:    nodePool1Path,
	}
	job2Bytes, _ = json.Marshal(&job2)

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
	service2Bytes, _ = json.Marshal(&service2)

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

	expected := NewResolutionTree()
	expected.Insert(tree.RootPath(), &ResolutionInfo{Component: job1})
	compareComponentTrees(t, "job", expected, result)

	// service
	result, err = r.Resolve(service1, system1ID, tree.RootPath(), nil, DepthInfinite)
	if err != nil {
		t.Errorf("expected no error resolving plain job, got :%v", err)
	}

	expected = NewResolutionTree()
	expected.Insert(tree.RootPath(), &ResolutionInfo{Component: service1})
	compareComponentTrees(t, "service", expected, result)

	// system
	result, err = r.Resolve(system1, system1ID, tree.RootPath(), nil, DepthInfinite)
	if err != nil {
		t.Errorf("expected no error resolving plain job, got :%v", err)
	}

	expected = NewResolutionTree()
	expected.Insert(tree.RootPath(), &ResolutionInfo{Component: system1})
	expected.Insert(tree.RootPath().Child("job"), &ResolutionInfo{Component: job1})
	expected.Insert(tree.RootPath().Child("service"), &ResolutionInfo{Component: service1})
	compareComponentTrees(t, "system", expected, result)
}

func TestComponentResolver_CommitReference(t *testing.T) {
	r := resolver()

	// seed repo
	repo := repoURL(repo1)
	os.RemoveAll(workDir)
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
	jobCommitStr := jobCommit.String()

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
	serviceCommitStr := serviceCommit.String()

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
	systemCommitStr := systemCommit.String()

	testResolutionSuccesses(
		t,
		r,
		[]successfulResolutionTest{
			{
				name: "job",
				c: &definitionv1.Reference{
					GitRepository: &definitionv1.GitRepositoryReference{
						GitRepository: &definitionv1.GitRepository{
							URL:    repo,
							Commit: &jobCommitStr,
						},
					},
				},
				depth: DepthInfinite,
				expected: map[tree.Path]*ResolutionInfo{
					tree.RootPath(): {
						Component: job1,
						Commit: &git.CommitReference{
							RepositoryURL: repo,
							Commit:        jobCommit.String(),
						},
					},
				},
			},
			{
				name: "service",
				c: &definitionv1.Reference{
					GitRepository: &definitionv1.GitRepositoryReference{
						GitRepository: &definitionv1.GitRepository{
							URL:    repo,
							Commit: &serviceCommitStr,
						},
					},
				},
				depth: DepthInfinite,
				expected: map[tree.Path]*ResolutionInfo{
					tree.RootPath(): {
						Component: service1,
						Commit: &git.CommitReference{
							RepositoryURL: repo,
							Commit:        serviceCommitStr,
						},
					},
				},
			},
			{
				name: "system",
				c: &definitionv1.Reference{
					GitRepository: &definitionv1.GitRepositoryReference{
						GitRepository: &definitionv1.GitRepository{
							URL:    repo,
							Commit: &systemCommitStr,
						},
					},
				},
				depth: DepthInfinite,
				expected: map[tree.Path]*ResolutionInfo{
					tree.RootPath(): {
						Component: system1,
						Commit: &git.CommitReference{
							RepositoryURL: repo,
							Commit:        systemCommitStr,
						},
					},
					tree.RootPath().Child("job"): {
						Component: job1,
						Commit: &git.CommitReference{
							RepositoryURL: repo,
							Commit:        systemCommitStr,
						},
					},
					tree.RootPath().Child("service"): {
						Component: service1,
						Commit: &git.CommitReference{
							RepositoryURL: repo,
							Commit:        systemCommitStr,
						},
					},
				},
			},
		},
	)

	invalidCommit := "0123456789abcdef0123456789abcdef01234567"
	testResolutionFailure(
		t,
		r,
		&failedResolutionTest{
			name: "invalid commit",
			c: &definitionv1.Reference{
				GitRepository: &definitionv1.GitRepositoryReference{
					GitRepository: &definitionv1.GitRepository{
						URL:    repo,
						Commit: &invalidCommit,
					},
				},
			},
		},
	)
}

func TestComponentResolver_BranchReference(t *testing.T) {
	r := resolver()

	// seed repo
	repo := repoURL(repo1)
	os.RemoveAll(workDir)
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

	testResolutionSuccess(
		t,
		r,
		&successfulResolutionTest{
			name: "initial branch commit",
			c: &definitionv1.Reference{
				GitRepository: &definitionv1.GitRepositoryReference{
					GitRepository: &definitionv1.GitRepository{
						URL:    repo,
						Branch: &devBranch,
					},
				},
			},
			depth: DepthInfinite,
			expected: map[tree.Path]*ResolutionInfo{
				tree.RootPath(): {
					Component: service1,
					Commit: &git.CommitReference{
						RepositoryURL: repo,
						Commit:        serviceCommit.String(),
					},
				},
			},
		},
	)

	// commit again and see the reference be updated
	job2Commit, err := git.WriteAndCommitFile(
		repo,
		DefaultFile,
		job2Bytes,
		0700,
		"job",
	)
	if err != nil {
		t.Fatalf("error commiting to repo: %v", err)
	}

	testResolutionSuccess(
		t,
		r,
		&successfulResolutionTest{
			name: "second branch commit",
			c: &definitionv1.Reference{
				GitRepository: &definitionv1.GitRepositoryReference{
					GitRepository: &definitionv1.GitRepository{
						URL:    repo,
						Branch: &devBranch,
					},
				},
			},
			depth: DepthInfinite,
			expected: map[tree.Path]*ResolutionInfo{
				tree.RootPath(): {
					Component: job2,
					Commit: &git.CommitReference{
						RepositoryURL: repo,
						Commit:        job2Commit.String(),
					},
				},
			},
		},
	)

	invalidBranch := "foo"
	testResolutionFailure(
		t,
		r,
		&failedResolutionTest{
			name: "invalid branch",
			c: &definitionv1.Reference{
				GitRepository: &definitionv1.GitRepositoryReference{
					GitRepository: &definitionv1.GitRepository{
						URL:    repo,
						Branch: &invalidBranch,
					},
				},
			},
		},
	)
}

func TestComponentResolver_TagReference(t *testing.T) {
	r := resolver()

	// seed repo
	repo := repoURL(repo1)
	os.RemoveAll(workDir)
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

	fooTag := "foo"
	err = git.Tag(repo, jobCommit, fooTag)
	if err != nil {
		t.Fatalf("error tagging repo: %v", err)
	}

	serviceCommit, err := git.WriteAndCommitFile(
		repo,
		DefaultFile,
		service1Bytes,
		0700,
		"service",
	)
	if err != nil {
		t.Fatalf("error commiting to repo: %v", err)
	}

	barTag := "bar"
	err = git.Tag(repo, serviceCommit, barTag)
	if err != nil {
		t.Fatalf("error tagging repo: %v", err)
	}

	testResolutionSuccesses(
		t,
		r,
		[]successfulResolutionTest{
			{
				name: "foo tag",
				c: &definitionv1.Reference{
					GitRepository: &definitionv1.GitRepositoryReference{
						GitRepository: &definitionv1.GitRepository{
							URL: repo,
							Tag: &fooTag,
						},
					},
				},
				depth: DepthInfinite,
				expected: map[tree.Path]*ResolutionInfo{
					tree.RootPath(): {
						Component: job1,
						Commit: &git.CommitReference{
							RepositoryURL: repo,
							Commit:        jobCommit.String(),
						},
					},
				},
			},
			{
				name: "bar tag",
				c: &definitionv1.Reference{
					GitRepository: &definitionv1.GitRepositoryReference{
						GitRepository: &definitionv1.GitRepository{
							URL: repo,
							Tag: &barTag,
						},
					},
				},
				depth: DepthInfinite,
				expected: map[tree.Path]*ResolutionInfo{
					tree.RootPath(): {
						Component: service1,
						Commit: &git.CommitReference{
							RepositoryURL: repo,
							Commit:        serviceCommit.String(),
						},
					},
				},
			},
		},
	)

	invalidTag := "invalid"
	testResolutionFailure(
		t,
		r,
		&failedResolutionTest{
			name: "invalid branch",
			c: &definitionv1.Reference{
				GitRepository: &definitionv1.GitRepositoryReference{
					GitRepository: &definitionv1.GitRepository{
						URL: repo,
						Tag: &invalidTag,
					},
				},
			},
		},
	)
}

func TestComponentResolver_VersionReference(t *testing.T) {
	r := resolver()

	// seed repo
	repo := repoURL(repo1)
	os.RemoveAll(workDir)
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

	err = git.Tag(repo, jobCommit, "1.0.0")
	if err != nil {
		t.Fatalf("error tagging repo: %v", err)
	}

	serviceCommit, err := git.WriteAndCommitFile(
		repo,
		DefaultFile,
		service1Bytes,
		0700,
		"service",
	)
	if err != nil {
		t.Fatalf("error commiting to repo: %v", err)
	}

	err = git.Tag(repo, serviceCommit, "2.0.0")
	if err != nil {
		t.Fatalf("error tagging repo: %v", err)
	}

	exactVersion := "1.0.0"
	patchVersion := "1.0.x"
	minorVersion := "1.x"

	testResolutionSuccesses(
		t,
		r,
		[]successfulResolutionTest{
			{
				name: "updated minor, exact ref",
				c: &definitionv1.Reference{
					GitRepository: &definitionv1.GitRepositoryReference{
						GitRepository: &definitionv1.GitRepository{
							URL:     repo,
							Version: &exactVersion,
						},
					},
				},
				depth: DepthInfinite,
				expected: map[tree.Path]*ResolutionInfo{
					tree.RootPath(): {
						Component: job1,
						Commit: &git.CommitReference{
							RepositoryURL: repo,
							Commit:        jobCommit.String(),
						},
					},
				},
			},
			{
				name: "updated minor, patch ref",
				c: &definitionv1.Reference{
					GitRepository: &definitionv1.GitRepositoryReference{
						GitRepository: &definitionv1.GitRepository{
							URL:     repo,
							Version: &patchVersion,
						},
					},
				},
				depth: DepthInfinite,
				expected: map[tree.Path]*ResolutionInfo{
					tree.RootPath(): {
						Component: job1,
						Commit: &git.CommitReference{
							RepositoryURL: repo,
							Commit:        jobCommit.String(),
						},
					},
				},
			},
			{
				name: "updated minor, minor ref",
				c: &definitionv1.Reference{
					GitRepository: &definitionv1.GitRepositoryReference{
						GitRepository: &definitionv1.GitRepository{
							URL:     repo,
							Version: &minorVersion,
						},
					},
				},
				depth: DepthInfinite,
				expected: map[tree.Path]*ResolutionInfo{
					tree.RootPath(): {
						Component: job1,
						Commit: &git.CommitReference{
							RepositoryURL: repo,
							Commit:        jobCommit.String(),
						},
					},
				},
			},
		},
	)

	job2Commit, err := git.WriteAndCommitFile(
		repo,
		DefaultFile,
		job2Bytes,
		0700,
		"job",
	)
	if err != nil {
		t.Fatalf("error commiting to repo: %v", err)
	}

	err = git.Tag(repo, job2Commit, "1.0.1")
	if err != nil {
		t.Fatalf("error tagging repo: %v", err)
	}

	testResolutionSuccesses(
		t,
		r,
		[]successfulResolutionTest{
			// exact version reference shouldn't have changed
			{
				name: "updated minor, exact ref",
				c: &definitionv1.Reference{
					GitRepository: &definitionv1.GitRepositoryReference{
						GitRepository: &definitionv1.GitRepository{
							URL:     repo,
							Version: &exactVersion,
						},
					},
				},
				depth: DepthInfinite,
				expected: map[tree.Path]*ResolutionInfo{
					tree.RootPath(): {
						Component: job1,
						Commit: &git.CommitReference{
							RepositoryURL: repo,
							Commit:        jobCommit.String(),
						},
					},
				},
			},
			// minor and patch versions reference should have changed
			{
				name: "updated minor, patch ref",
				c: &definitionv1.Reference{
					GitRepository: &definitionv1.GitRepositoryReference{
						GitRepository: &definitionv1.GitRepository{
							URL:     repo,
							Version: &patchVersion,
						},
					},
				},
				depth: DepthInfinite,
				expected: map[tree.Path]*ResolutionInfo{
					tree.RootPath(): {
						Component: job2,
						Commit: &git.CommitReference{
							RepositoryURL: repo,
							Commit:        job2Commit.String(),
						},
					},
				},
			},
			{
				name: "updated minor, minor ref",
				c: &definitionv1.Reference{
					GitRepository: &definitionv1.GitRepositoryReference{
						GitRepository: &definitionv1.GitRepository{
							URL:     repo,
							Version: &minorVersion,
						},
					},
				},
				depth: DepthInfinite,
				expected: map[tree.Path]*ResolutionInfo{
					tree.RootPath(): {
						Component: job2,
						Commit: &git.CommitReference{
							RepositoryURL: repo,
							Commit:        job2Commit.String(),
						},
					},
				},
			},
		},
	)

	service2Commit, err := git.WriteAndCommitFile(
		repo,
		DefaultFile,
		service2Bytes,
		0700,
		"job",
	)
	if err != nil {
		t.Fatalf("error commiting to repo: %v", err)
	}

	err = git.Tag(repo, service2Commit, "1.1.0")
	if err != nil {
		t.Fatalf("error tagging repo: %v", err)
	}

	testResolutionSuccesses(
		t,
		r,
		[]successfulResolutionTest{
			// exact and patch version references shouldn't have changed
			{
				name: "updated minor, exact ref",
				c: &definitionv1.Reference{
					GitRepository: &definitionv1.GitRepositoryReference{
						GitRepository: &definitionv1.GitRepository{
							URL:     repo,
							Version: &exactVersion,
						},
					},
				},
				depth: DepthInfinite,
				expected: map[tree.Path]*ResolutionInfo{
					tree.RootPath(): {
						Component: job1,
						Commit: &git.CommitReference{
							RepositoryURL: repo,
							Commit:        jobCommit.String(),
						},
					},
				},
			},
			{
				name: "updated minor, patch ref",
				c: &definitionv1.Reference{
					GitRepository: &definitionv1.GitRepositoryReference{
						GitRepository: &definitionv1.GitRepository{
							URL:     repo,
							Version: &patchVersion,
						},
					},
				},
				depth: DepthInfinite,
				expected: map[tree.Path]*ResolutionInfo{
					tree.RootPath(): {
						Component: job2,
						Commit: &git.CommitReference{
							RepositoryURL: repo,
							Commit:        job2Commit.String(),
						},
					},
				},
			},
			// minor version reference should have changed
			{
				name: "updated minor, minor ref",
				c: &definitionv1.Reference{
					GitRepository: &definitionv1.GitRepositoryReference{
						GitRepository: &definitionv1.GitRepository{
							URL:     repo,
							Version: &minorVersion,
						},
					},
				},
				depth: DepthInfinite,
				expected: map[tree.Path]*ResolutionInfo{
					tree.RootPath(): {
						Component: service2,
						Commit: &git.CommitReference{
							RepositoryURL: repo,
							Commit:        service2Commit.String(),
						},
					},
				},
			},
		},
	)

	invalidVersion := "foo"
	testResolutionFailure(
		t,
		r,
		&failedResolutionTest{
			name: "invalid branch",
			c: &definitionv1.Reference{
				GitRepository: &definitionv1.GitRepositoryReference{
					GitRepository: &definitionv1.GitRepository{
						URL:     repo,
						Version: &invalidVersion,
					},
				},
			},
		},
	)
}

func TestComponentResolver_FileReference(t *testing.T) {
	r := resolver()

	// seed repo
	repo := repoURL(repo1)
	os.RemoveAll(workDir)
	err := git.Init(repo)
	if err != nil {
		t.Fatalf("error initializing repo: %v", err)
	}

	jobFile := "job1.json"
	err = git.WriteAndAddFile(
		repo,
		jobFile,
		job1Bytes,
		0700,
	)
	if err != nil {
		t.Fatalf("error adding file to repo: %v", err)
	}

	serviceFile := "service1.json"
	err = git.WriteAndAddFile(
		repo,
		serviceFile,
		service1Bytes,
		0700,
	)
	if err != nil {
		t.Fatalf("error adding file to repo: %v", err)
	}

	commit, err := git.Commit(repo, "commit")
	if err != nil {
		t.Fatalf("error commiting file to repo: %v", err)
	}

	commitStr := commit.String()
	testResolutionSuccesses(
		t,
		r,
		[]successfulResolutionTest{
			// resolving a git reference with a specific file
			{
				name: "job file",
				c: &definitionv1.Reference{
					GitRepository: &definitionv1.GitRepositoryReference{
						GitRepository: &definitionv1.GitRepository{
							URL:    repo,
							Commit: &commitStr,
						},
						File: &jobFile,
					},
				},
				depth: DepthInfinite,
				expected: map[tree.Path]*ResolutionInfo{
					tree.RootPath(): {
						Component: job1,
						Commit: &git.CommitReference{
							RepositoryURL: repo,
							Commit:        commitStr,
						},
					},
				},
			},
			{
				name: "service file",
				c: &definitionv1.Reference{
					GitRepository: &definitionv1.GitRepositoryReference{
						GitRepository: &definitionv1.GitRepository{
							URL:    repo,
							Commit: &commitStr,
						},
						File: &serviceFile,
					},
				},
				depth: DepthInfinite,
				expected: map[tree.Path]*ResolutionInfo{
					tree.RootPath(): {
						Component: service1,
						Commit: &git.CommitReference{
							RepositoryURL: repo,
							Commit:        commitStr,
						},
					},
				},
			},
			// resolving a local file reference with context
			{
				name: "job file",
				c: &definitionv1.Reference{
					File: &jobFile,
				},
				ctx: &git.CommitReference{
					RepositoryURL: repo,
					Commit:        commitStr,
				},
				depth: DepthInfinite,
				expected: map[tree.Path]*ResolutionInfo{
					tree.RootPath(): {
						Component: job1,
						Commit: &git.CommitReference{
							RepositoryURL: repo,
							Commit:        commitStr,
						},
					},
				},
			},
			{
				name: "service file",
				c: &definitionv1.Reference{
					File: &serviceFile,
				},
				ctx: &git.CommitReference{
					RepositoryURL: repo,
					Commit:        commitStr,
				},
				depth: DepthInfinite,
				expected: map[tree.Path]*ResolutionInfo{
					tree.RootPath(): {
						Component: service1,
						Commit: &git.CommitReference{
							RepositoryURL: repo,
							Commit:        commitStr,
						},
					},
				},
			},
		},
	)

	invalidFile := DefaultFile
	invalidCommit := "0123456789abcdef0123456789abcdef01234567"
	testResolutionFailures(
		t,
		r,
		[]failedResolutionTest{
			{
				name: "invalid file with valid git reference",
				c: &definitionv1.Reference{
					GitRepository: &definitionv1.GitRepositoryReference{
						GitRepository: &definitionv1.GitRepository{
							URL:    repo,
							Commit: &commitStr,
						},
						File: &invalidFile,
					},
				},
			},
			{
				name: "invalid file with valid git context",
				c: &definitionv1.Reference{
					File: &invalidFile,
				},
				ctx: &git.CommitReference{
					RepositoryURL: repo,
					Commit:        commitStr,
				},
			},
			{
				name: "valid file with invalid git commit",
				c: &definitionv1.Reference{
					GitRepository: &definitionv1.GitRepositoryReference{
						GitRepository: &definitionv1.GitRepository{
							URL:    repo,
							Commit: &invalidCommit,
						},
						File: &serviceFile,
					},
				},
			},
			{
				name: "valid file with invalid git repo",
				c: &definitionv1.Reference{
					GitRepository: &definitionv1.GitRepositoryReference{
						GitRepository: &definitionv1.GitRepository{
							URL:    repoURL("invalid"),
							Commit: &commitStr,
						},
						File: &serviceFile,
					},
				},
			},
			{
				name: "valid file with invalid git context",
				c: &definitionv1.Reference{
					File: &serviceFile,
				},
				ctx: &git.CommitReference{
					RepositoryURL: repo,
					Commit:        invalidCommit,
				},
			},
		},
	)
}

type successfulResolutionTest struct {
	name     string
	c        component.Interface
	p        tree.Path
	ctx      *git.CommitReference
	depth    int
	expected map[tree.Path]*ResolutionInfo
}

func testResolutionSuccesses(t *testing.T, r ComponentResolver, tests []successfulResolutionTest) {
	for _, test := range tests {
		testResolutionSuccess(t, r, &test)
	}
}

func testResolutionSuccess(t *testing.T, r ComponentResolver, test *successfulResolutionTest) {
	if test.p == "" {
		test.p = tree.RootPath()
	}
	result, err := r.Resolve(test.c, system1ID, test.p, test.ctx, test.depth)
	if err != nil {
		t.Errorf("expected no error resolving %v, got :%v", test.name, err)
	}

	e := NewResolutionTree()
	for p, i := range test.expected {
		e.Insert(p, i)
	}
	compareComponentTrees(t, test.name, e, result)
}

type failedResolutionTest struct {
	name string
	c    component.Interface
	p    tree.Path
	ctx  *git.CommitReference
	// TODO(kevindrosendahl): add expected error once errors are classified
}

func testResolutionFailures(t *testing.T, r ComponentResolver, tests []failedResolutionTest) {
	for _, test := range tests {
		testResolutionFailure(t, r, &test)
	}
}

func testResolutionFailure(t *testing.T, r ComponentResolver, test *failedResolutionTest) {
	if test.p == "" {
		test.p = tree.RootPath()
	}
	_, err := r.Resolve(test.c, system1ID, test.p, test.ctx, DepthInfinite)
	if err == nil {
		t.Errorf("expected error resolving %v but got none", test.name)
	}
}

func compareComponentTrees(t *testing.T, name string, expected, actual *ResolutionTree) {
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
