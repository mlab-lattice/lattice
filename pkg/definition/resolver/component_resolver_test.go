package resolver

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"testing"

	gitplumbing "gopkg.in/src-d/go-git.v4/plumbing"

	"github.com/stretchr/testify/require"

	definitionv1 "github.com/mlab-lattice/lattice/pkg/definition/v1"

	"github.com/mlab-lattice/lattice/pkg/api/v1"
	"github.com/mlab-lattice/lattice/pkg/definition/component"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	"github.com/mlab-lattice/lattice/pkg/util/git"
)

const workDir = "/tmp/lattice/test/pkg/definition/resolver/component_resolver"

var (
	systemID   v1.SystemID = "foo"
	remote1Dir             = fmt.Sprintf("%v/remote1", workDir)

	service1 = definitionv1.Service{
		Container: definitionv1.Container{
			Exec: &definitionv1.ContainerExec{
				Command: []string{"foo"},
			},
		},
	}
	service2 = definitionv1.Service{
		Container: definitionv1.Container{
			Exec: &definitionv1.ContainerExec{
				Command: []string{"bar"},
			},
		},
	}
	service3 = definitionv1.Service{
		Container: definitionv1.Container{
			Exec: &definitionv1.ContainerExec{
				Command: []string{"baz"},
			},
		},
	}
)

// TODO: add tests for relative paths
func TestReferenceResolver(t *testing.T) {
	testFileReferenceResolve(t)
	testGitReferenceResolve(t)
}

func testFileReferenceResolve(t *testing.T) {
	cleanReferenceResolverWorkDir(t)

	if err := git.Init(remote1Dir); err != nil {
		t.Fatal(err)
	}

	serviceBytes, err := json.Marshal(&service1)
	if err != nil {
		t.Fatal(err)
	}

	servicePath := "service.json"
	commit, err := git.WriteAndCommitFile(remote1Dir, servicePath, serviceBytes, 0700, "my commit")
	if err != nil {
		t.Fatal(err)
	}

	r, err := NewComponentResolver(workDir, true, NewMemoryTemplateStore(), NewMemorySecretStore())
	if err != nil {
		t.Fatal(err)
	}

	ctx := &git.FileReference{
		CommitReference: git.CommitReference{
			RepositoryURL: fmt.Sprintf("file://%v", remote1Dir),
			Commit:        commit.String(),
		},
		File: "foo",
	}

	ref := &definitionv1.Reference{
		File: &servicePath,
	}

	if err := shouldResolveToServiceCtx(r, ctx, ref, &service1); err != nil {
		t.Error(err)
	}

	bar := "bar"
	ref.File = &bar
	if err := shouldFailToResolveCtx(r, ctx, ref); err != nil {
		t.Error(err)
	}
}

func testGitReferenceResolve(t *testing.T) {
	testCommitGitReferenceResolve(t)
	testBranchGitReferenceResolve(t)
	testTagAndVersionGitReferenceResolve(t)
}

func testCommitGitReferenceResolve(t *testing.T) {
	cleanReferenceResolverWorkDir(t)

	if err := git.Init(remote1Dir); err != nil {
		t.Fatal(err)
	}

	serviceBytes, err := json.Marshal(&service1)
	if err != nil {
		t.Fatal(err)
	}

	service2Bytes, err := json.Marshal(&service2)
	if err != nil {
		t.Fatal(err)
	}

	servicePath := "service.json"
	commit, err := git.WriteAndCommitFile(remote1Dir, servicePath, serviceBytes, 0700, "my commit")
	if err != nil {
		t.Fatal(err)
	}

	defaultCommit, err := git.WriteAndCommitFile(remote1Dir, DefaultFile, service2Bytes, 0700, "default file")
	if err != nil {
		t.Fatal(err)
	}

	r, err := NewComponentResolver(workDir, true, NewMemoryTemplateStore(), NewMemorySecretStore())
	if err != nil {
		t.Fatal(err)
	}

	foo := "foo"

	tests := []struct {
		Description string
		Test        func() error
	}{
		{
			Description: "valid commit",
			Test: func() error {
				commitStr := commit.String()
				ref := &definitionv1.Reference{
					GitRepository: &definitionv1.GitRepositoryReference{
						GitRepository: &definitionv1.GitRepository{
							URL:    fmt.Sprintf("file://%v", remote1Dir),
							Commit: &commitStr,
						},
						File: &servicePath,
					},
				}

				return shouldResolveToService(r, ref, &service1)
			},
		},
		{
			Description: "default file",
			Test: func() error {
				commitStr := defaultCommit.String()
				ref := &definitionv1.Reference{
					GitRepository: &definitionv1.GitRepositoryReference{
						GitRepository: &definitionv1.GitRepository{
							URL:    fmt.Sprintf("file://%v", remote1Dir),
							Commit: &commitStr,
						},
					},
				}

				return shouldResolveToService(r, ref, &service2)
			},
		},
		{
			Description: "invalid file",
			Test: func() error {
				commitStr := commit.String()
				ref := &definitionv1.Reference{
					GitRepository: &definitionv1.GitRepositoryReference{
						GitRepository: &definitionv1.GitRepository{
							URL:    fmt.Sprintf("file://%v", remote1Dir),
							Commit: &commitStr,
						},
						File: &foo,
					},
				}

				return shouldFailToResolve(r, ref)
			},
		},
		{
			Description: "invalid git commit",
			Test: func() error {
				commitStr := "0123456789012345678901234567890123456789"
				ref := &definitionv1.Reference{
					GitRepository: &definitionv1.GitRepositoryReference{
						GitRepository: &definitionv1.GitRepository{
							URL:    fmt.Sprintf("file://%v", remote1Dir),
							Commit: &commitStr,
						},
						File: &servicePath,
					},
				}

				return shouldFailToResolve(r, ref)
			},
		},
	}

	for _, test := range tests {
		if err := test.Test(); err != nil {
			t.Errorf("error testing %v: %v", test.Description, err)
		}
	}
}

func testBranchGitReferenceResolve(t *testing.T) {
	cleanReferenceResolverWorkDir(t)

	if err := git.Init(remote1Dir); err != nil {
		t.Fatal(err)
	}

	serviceBytes, err := json.Marshal(&service1)
	if err != nil {
		t.Fatal(err)
	}

	servicePath := "service.json"
	commit, err := git.WriteAndCommitFile(remote1Dir, servicePath, serviceBytes, 0700, "my commit")
	if err != nil {
		t.Fatal(err)
	}

	branchName := "foo"
	if err := git.CreateBranch(remote1Dir, branchName, commit); err != nil {
		t.Fatal(err)
	}

	r, err := NewComponentResolver(workDir, true, NewMemoryTemplateStore(), NewMemorySecretStore())
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		Description string
		Test        func() error
	}{
		{
			Description: "valid branch",
			Test: func() error {
				ref := &definitionv1.Reference{
					GitRepository: &definitionv1.GitRepositoryReference{
						GitRepository: &definitionv1.GitRepository{
							URL:    fmt.Sprintf("file://%v", remote1Dir),
							Branch: &branchName,
						},
						File: &servicePath,
					},
				}

				return shouldResolveToService(r, ref, &service1)
			},
		},
		{
			Description: "different branch",
			Test: func() error {
				branchName := "bar"
				if err := git.CreateBranch(remote1Dir, branchName, commit); err != nil {
					t.Fatal(err)
				}

				serviceBytes, err := json.Marshal(&service2)
				if err != nil {
					t.Fatal(err)
				}

				servicePath := "service.json"
				_, err = git.WriteAndCommitFile(remote1Dir, servicePath, serviceBytes, 0700, "my commit")
				if err != nil {
					t.Fatal(err)
				}

				ref := &definitionv1.Reference{
					GitRepository: &definitionv1.GitRepositoryReference{
						GitRepository: &definitionv1.GitRepository{
							URL:    fmt.Sprintf("file://%v", remote1Dir),
							Branch: &branchName,
						},
						File: &servicePath,
					},
				}

				return shouldResolveToService(r, ref, &service2)
			},
		},
		{
			Description: "update branch",
			Test: func() error {
				branchName := "bar"
				serviceBytes, err := json.Marshal(&service3)
				if err != nil {
					t.Fatal(err)
				}

				servicePath := "service.json"
				_, err = git.WriteAndCommitFile(remote1Dir, servicePath, serviceBytes, 0700, "my commit")
				if err != nil {
					t.Fatal(err)
				}

				ref := &definitionv1.Reference{
					GitRepository: &definitionv1.GitRepositoryReference{
						GitRepository: &definitionv1.GitRepository{
							URL:    fmt.Sprintf("file://%v", remote1Dir),
							Branch: &branchName,
						},
						File: &servicePath,
					},
				}

				return shouldResolveToService(r, ref, &service3)
			},
		},
		{
			Description: "invalid file",
			Test: func() error {
				branchName := "bar"
				foo := "foo"
				ref := &definitionv1.Reference{
					GitRepository: &definitionv1.GitRepositoryReference{
						GitRepository: &definitionv1.GitRepository{
							URL:    fmt.Sprintf("file://%v", remote1Dir),
							Branch: &branchName,
						},
						File: &foo,
					},
				}

				return shouldFailToResolve(r, ref)
			},
		},
		{
			Description: "invalid branch",
			Test: func() error {
				branch := "invalid"
				ref := &definitionv1.Reference{
					GitRepository: &definitionv1.GitRepositoryReference{
						GitRepository: &definitionv1.GitRepository{
							URL:    fmt.Sprintf("file://%v", remote1Dir),
							Branch: &branch,
						},
						File: &servicePath,
					},
				}

				return shouldFailToResolve(r, ref)
			},
		},
	}

	for _, test := range tests {
		if err := test.Test(); err != nil {
			t.Errorf("error testing %v: %v", test.Description, err)
		}
	}
}

func testTagAndVersionGitReferenceResolve(t *testing.T) {
	cleanReferenceResolverWorkDir(t)

	if err := git.Init(remote1Dir); err != nil {
		t.Fatal(err)
	}

	serviceBytes, err := json.Marshal(&service1)
	if err != nil {
		t.Fatal(err)
	}

	servicePath := "service.json"
	commit1, err := git.WriteAndCommitFile(remote1Dir, servicePath, serviceBytes, 0700, "my commit1")
	if err != nil {
		t.Fatal(err)
	}

	r, err := NewComponentResolver(workDir, true, NewMemoryTemplateStore(), NewMemorySecretStore())
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		Description string
		Test        func() error
	}{
		{
			Description: "valid tag",
			Test: func() error {
				tagName := "foo"
				if err := git.Tag(remote1Dir, commit1, tagName); err != nil {
					t.Fatal(err)
				}

				ref := &definitionv1.Reference{
					GitRepository: &definitionv1.GitRepositoryReference{
						GitRepository: &definitionv1.GitRepository{
							URL: fmt.Sprintf("file://%v", remote1Dir),
							Tag: &tagName,
						},
						File: &servicePath,
					},
				}

				return shouldResolveToService(r, ref, &service1)
			},
		},
		{
			Description: "strict semver patch should initially resolve",
			Test: func() error {
				// minor and patch semver should work initially
				tagName := "1.0.0"
				if err := git.Tag(remote1Dir, commit1, tagName); err != nil {
					t.Fatal(err)
				}

				patchSemverTag := "1.0.x"
				ref := &definitionv1.Reference{
					GitRepository: &definitionv1.GitRepositoryReference{
						GitRepository: &definitionv1.GitRepository{
							URL:     fmt.Sprintf("file://%v", remote1Dir),
							Version: &patchSemverTag,
						},
						File: &servicePath,
					},
				}

				return shouldResolveToService(r, ref, &service1)
			},
		},
		{
			Description: "strict semver minor should initially resolve",
			Test: func() error {
				minorSemverTag := "1.x"
				ref := &definitionv1.Reference{
					GitRepository: &definitionv1.GitRepositoryReference{
						GitRepository: &definitionv1.GitRepository{
							URL:     fmt.Sprintf("file://%v", remote1Dir),
							Version: &minorSemverTag,
						},
						File: &servicePath,
					},
				}

				return shouldResolveToService(r, ref, &service1)
			},
		},
		{
			Description: "strict semver invalid major should not initially resolve",
			Test: func() error {
				invalidSemverTag := "2.x"
				ref := &definitionv1.Reference{
					GitRepository: &definitionv1.GitRepositoryReference{
						GitRepository: &definitionv1.GitRepository{
							URL:     fmt.Sprintf("file://%v", remote1Dir),
							Version: &invalidSemverTag,
						},
						File: &servicePath,
					},
				}

				return shouldFailToResolve(r, ref)
			},
		},
		{
			Description: "strict semver patch resolve should update with patch update",
			Test: func() error {
				// minor and patch semver should resolve new definition with a patch bump
				serviceBytes, err := json.Marshal(&service2)
				if err != nil {
					t.Fatal(err)
				}

				servicePath := "service.json"
				commit2, err := git.WriteAndCommitFile(remote1Dir, servicePath, serviceBytes, 0700, "my commit1")
				if err != nil {
					t.Fatal(err)
				}

				tagName := "1.0.1"
				if err := git.Tag(remote1Dir, commit2, tagName); err != nil {
					t.Fatal(err)
				}

				patchTag := "1.0.x"
				ref := &definitionv1.Reference{
					GitRepository: &definitionv1.GitRepositoryReference{
						GitRepository: &definitionv1.GitRepository{
							URL:     fmt.Sprintf("file://%v", remote1Dir),
							Version: &patchTag,
						},
						File: &servicePath,
					},
				}

				return shouldResolveToService(r, ref, &service2)
			},
		},
		{
			Description: "strict semver minor resolve should update with patch update",
			Test: func() error {
				minorTag := "1.x"
				ref := &definitionv1.Reference{
					GitRepository: &definitionv1.GitRepositoryReference{
						GitRepository: &definitionv1.GitRepository{
							URL:     fmt.Sprintf("file://%v", remote1Dir),
							Version: &minorTag,
						},
						File: &servicePath,
					},
				}

				return shouldResolveToService(r, ref, &service2)
			},
		},
		{
			Description: "strict semver patch resolve should not update with minor update",
			Test: func() error {
				serviceBytes, err := json.Marshal(&service3)
				if err != nil {
					t.Fatal(err)
				}

				servicePath := "service.json"
				commit, err := git.WriteAndCommitFile(remote1Dir, servicePath, serviceBytes, 0700, "my commit1")
				if err != nil {
					t.Fatal(err)
				}

				tagName := "1.1.0"
				if err := git.Tag(remote1Dir, commit, tagName); err != nil {
					t.Fatal(err)
				}

				patchSemverTag := "1.0.x"
				ref := &definitionv1.Reference{
					GitRepository: &definitionv1.GitRepositoryReference{
						GitRepository: &definitionv1.GitRepository{
							URL:     fmt.Sprintf("file://%v", remote1Dir),
							Version: &patchSemverTag,
						},
						File: &servicePath,
					},
				}

				return shouldResolveToService(r, ref, &service2)
			},
		},
		{
			Description: "strict semver minor resolve should update with minor update",
			Test: func() error {
				minorSemverTag := "1.x"
				ref := &definitionv1.Reference{
					GitRepository: &definitionv1.GitRepositoryReference{
						GitRepository: &definitionv1.GitRepository{
							URL:     fmt.Sprintf("file://%v", remote1Dir),
							Version: &minorSemverTag,
						},
						File: &servicePath,
					},
				}

				return shouldResolveToService(r, ref, &service3)
			},
		},
		{
			Description: "strict semver patch resolve should not update with major update",
			Test: func() error {
				serviceBytes, err := json.Marshal(&service3)
				if err != nil {
					t.Fatal(err)
				}

				servicePath := "service.json"
				commit, err := git.WriteAndCommitFile(remote1Dir, servicePath, serviceBytes, 0700, "my commit1")
				if err != nil {
					t.Fatal(err)
				}

				tagName := "2.0.0"
				if err := git.Tag(remote1Dir, commit, tagName); err != nil {
					t.Fatal(err)
				}

				patchSemverTag := "1.0.x"
				ref := &definitionv1.Reference{
					GitRepository: &definitionv1.GitRepositoryReference{
						GitRepository: &definitionv1.GitRepository{
							URL:     fmt.Sprintf("file://%v", remote1Dir),
							Version: &patchSemverTag,
						},
						File: &servicePath,
					},
				}

				return shouldResolveToService(r, ref, &service2)
			},
		},
		{
			Description: "strict semver minor resolve should not update with major update",
			Test: func() error {
				minorSemverTag := "1.x"
				ref := &definitionv1.Reference{
					GitRepository: &definitionv1.GitRepositoryReference{
						GitRepository: &definitionv1.GitRepository{
							URL:     fmt.Sprintf("file://%v", remote1Dir),
							Version: &minorSemverTag,
						},
						File: &servicePath,
					},
				}

				return shouldResolveToService(r, ref, &service3)
			},
		},
	}

	for _, test := range tests {
		if err := test.Test(); err != nil {
			t.Errorf("error testing %v: %v", test.Description, err)
		}
	}
}

var (
	dockerFile = `From ubuntu:18.04
Maintainer Anna Domino <anna@example.com>
`

	locationCommitStr            = "0000000000000000000000000000000000000000"
	serviceDockerBuildInheritAll = definitionv1.Service{
		Container: definitionv1.Container{
			Exec: &definitionv1.ContainerExec{
				Command: []string{"foo"},
			},
			Build: &definitionv1.ContainerBuild{
				DockerBuild: &definitionv1.DockerBuild{},
			},
		},
	}
	serviceDockerBuildInheritBuildContext = definitionv1.Service{
		Container: definitionv1.Container{
			Exec: &definitionv1.ContainerExec{
				Command: []string{"bar"},
			},
			Build: &definitionv1.ContainerBuild{
				DockerBuild: &definitionv1.DockerBuild{
					DockerFile: &definitionv1.DockerFile{
						Location: &definitionv1.Location{
							GitRepository: &definitionv1.GitRepository{
								URL:    fmt.Sprintf("file://%v", remote1Dir),
								Commit: &locationCommitStr,
							},
						},
					},
				},
			},
		},
	}
	serviceDockerBuildInheritDockerFile = definitionv1.Service{
		Container: definitionv1.Container{
			Exec: &definitionv1.ContainerExec{
				Command: []string{"baz"},
			},
			Build: &definitionv1.ContainerBuild{
				DockerBuild: &definitionv1.DockerBuild{
					BuildContext: &definitionv1.DockerBuildContext{
						Location: &definitionv1.Location{
							GitRepository: &definitionv1.GitRepository{
								URL:    fmt.Sprintf("file://%v", remote1Dir),
								Commit: &locationCommitStr,
							},
						},
					},
				},
			},
		},
	}
	serviceDockerBuildInheritNothing = definitionv1.Service{
		Container: definitionv1.Container{
			Exec: &definitionv1.ContainerExec{
				Command: []string{"yaz"},
			},
			Build: &definitionv1.ContainerBuild{
				DockerBuild: &definitionv1.DockerBuild{
					DockerFile: &definitionv1.DockerFile{
						Location: &definitionv1.Location{
							GitRepository: &definitionv1.GitRepository{
								URL:    fmt.Sprintf("file://%v", remote1Dir),
								Commit: &locationCommitStr,
							},
						},
					},
					BuildContext: &definitionv1.DockerBuildContext{
						Location: &definitionv1.Location{
							GitRepository: &definitionv1.GitRepository{
								URL:    fmt.Sprintf("file://%v", remote1Dir),
								Commit: &locationCommitStr,
							},
						},
					},
				},
			},
		},
	}
)

func TestDockerBuildHydation(t *testing.T) {
	testDockerBuildHydration(t)
}

func testDockerBuildHydration(t *testing.T) {
	var commit gitplumbing.Hash
	var r ComponentResolver

	servicePath := "service.json"

	setup := func(serviceBytes []byte) {
		cleanReferenceResolverWorkDir(t)

		err := git.Init(remote1Dir)
		if err != nil {
			t.Fatal(err)
		}

		commit, err = git.WriteAndCommitFile(remote1Dir, servicePath, serviceBytes, 0700, "my commit")
		if err != nil {
			t.Fatal(err)
		}

		r, err = NewComponentResolver(workDir, true, NewMemoryTemplateStore(), NewMemorySecretStore())
		if err != nil {
			t.Fatal(err)
		}
	}

	serviceDockerBuildInheritAllBytes, err := json.Marshal(&serviceDockerBuildInheritAll)
	if err != nil {
		t.Fatal(err)
	}
	serviceDockerBuildInheritBuildContextBytes, err := json.Marshal(&serviceDockerBuildInheritBuildContext)
	if err != nil {
		t.Fatal(err)
	}
	serviceDockerBuildInheritDockerFileBytes, err := json.Marshal(&serviceDockerBuildInheritDockerFile)
	if err != nil {
		t.Fatal(err)
	}
	serviceDockerBuildInheritNothingBytes, err := json.Marshal(&serviceDockerBuildInheritNothing)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		Description  string
		ServiceBytes []byte
		Test         func(*testing.T)
	}{
		{
			Description:  "Docker build inherit all",
			ServiceBytes: serviceDockerBuildInheritAllBytes,
			Test: func(t *testing.T) {
				commitStr := commit.String()
				ctx := &git.FileReference{
					CommitReference: git.CommitReference{
						RepositoryURL: fmt.Sprintf("file://%v", remote1Dir),
						Commit:        commitStr,
					},
					File: "foo",
				}

				ref := &definitionv1.Reference{
					GitRepository: &definitionv1.GitRepositoryReference{
						GitRepository: &definitionv1.GitRepository{
							URL:    fmt.Sprintf("file://%v", remote1Dir),
							Commit: &commitStr,
						},
						File: &servicePath,
					},
				}
				c, err := resolveReference(r, systemID, tree.RootPath(), ctx, ref, true)

				require.Nil(t, err)
				require.IsType(t, &definitionv1.Service{}, c)

				service := c.(*definitionv1.Service)

				require.Equal(t, &definitionv1.ContainerBuild{
					DockerBuild: &definitionv1.DockerBuild{
						DockerFile: &definitionv1.DockerFile{
							Location: &definitionv1.Location{
								GitRepository: &definitionv1.GitRepository{
									URL:    fmt.Sprintf("file://%v", remote1Dir),
									Commit: &commitStr,
								},
							},
							Path: definitionv1.DockerBuildDefaultPath,
						},
						BuildContext: &definitionv1.DockerBuildContext{
							Location: &definitionv1.Location{
								GitRepository: &definitionv1.GitRepository{
									URL:    fmt.Sprintf("file://%v", remote1Dir),
									Commit: &commitStr,
								},
							},
							Path: definitionv1.DockerBuildDefaultPath,
						},
					},
				}, service.Container.Build)
			},
		},
		{
			Description:  "Docker build inherit build context",
			ServiceBytes: serviceDockerBuildInheritBuildContextBytes,
			Test: func(t *testing.T) {
				commitStr := commit.String()
				ctx := &git.FileReference{
					CommitReference: git.CommitReference{
						RepositoryURL: fmt.Sprintf("file://%v", remote1Dir),
						Commit:        commitStr,
					},
					File: "foo",
				}

				ref := &definitionv1.Reference{
					GitRepository: &definitionv1.GitRepositoryReference{
						GitRepository: &definitionv1.GitRepository{
							URL:    fmt.Sprintf("file://%v", remote1Dir),
							Commit: &commitStr,
						},
						File: &servicePath,
					},
				}
				c, err := resolveReference(r, systemID, tree.RootPath(), ctx, ref, true)

				require.Nil(t, err)
				require.IsType(t, &definitionv1.Service{}, c)

				service := c.(*definitionv1.Service)

				require.Equal(t, &definitionv1.ContainerBuild{
					DockerBuild: &definitionv1.DockerBuild{
						DockerFile: &definitionv1.DockerFile{
							Location: &definitionv1.Location{
								GitRepository: &definitionv1.GitRepository{
									URL:    fmt.Sprintf("file://%v", remote1Dir),
									Commit: &locationCommitStr,
								},
							},
							Path: definitionv1.DockerBuildDefaultPath,
						},
						BuildContext: &definitionv1.DockerBuildContext{
							Location: &definitionv1.Location{
								GitRepository: &definitionv1.GitRepository{
									URL:    fmt.Sprintf("file://%v", remote1Dir),
									Commit: &commitStr,
								},
							},
							Path: definitionv1.DockerBuildDefaultPath,
						},
					},
				}, service.Container.Build)
			},
		},
		{
			Description:  "Docker build inherit docker file",
			ServiceBytes: serviceDockerBuildInheritDockerFileBytes,
			Test: func(t *testing.T) {
				commitStr := commit.String()
				ctx := &git.FileReference{
					CommitReference: git.CommitReference{
						RepositoryURL: fmt.Sprintf("file://%v", remote1Dir),
						Commit:        commitStr,
					},
					File: "foo",
				}

				ref := &definitionv1.Reference{
					GitRepository: &definitionv1.GitRepositoryReference{
						GitRepository: &definitionv1.GitRepository{
							URL:    fmt.Sprintf("file://%v", remote1Dir),
							Commit: &commitStr,
						},
						File: &servicePath,
					},
				}
				c, err := resolveReference(r, systemID, tree.RootPath(), ctx, ref, true)

				require.Nil(t, err)
				require.IsType(t, &definitionv1.Service{}, c)

				service := c.(*definitionv1.Service)

				require.Equal(t, &definitionv1.ContainerBuild{
					DockerBuild: &definitionv1.DockerBuild{
						DockerFile: &definitionv1.DockerFile{
							Location: &definitionv1.Location{
								GitRepository: &definitionv1.GitRepository{
									URL:    fmt.Sprintf("file://%v", remote1Dir),
									Commit: &commitStr,
								},
							},
							Path: definitionv1.DockerBuildDefaultPath,
						},
						BuildContext: &definitionv1.DockerBuildContext{
							Location: &definitionv1.Location{
								GitRepository: &definitionv1.GitRepository{
									URL:    fmt.Sprintf("file://%v", remote1Dir),
									Commit: &locationCommitStr,
								},
							},
							Path: definitionv1.DockerBuildDefaultPath,
						},
					},
				}, service.Container.Build)
			},
		},
		{
			Description:  "Docker build inherit nothing",
			ServiceBytes: serviceDockerBuildInheritNothingBytes,
			Test: func(t *testing.T) {
				commitStr := commit.String()
				ctx := &git.FileReference{
					CommitReference: git.CommitReference{
						RepositoryURL: fmt.Sprintf("file://%v", remote1Dir),
						Commit:        commitStr,
					},
					File: "foo",
				}

				ref := &definitionv1.Reference{
					GitRepository: &definitionv1.GitRepositoryReference{
						GitRepository: &definitionv1.GitRepository{
							URL:    fmt.Sprintf("file://%v", remote1Dir),
							Commit: &commitStr,
						},
						File: &servicePath,
					},
				}
				c, err := resolveReference(r, systemID, tree.RootPath(), ctx, ref, true)

				require.Nil(t, err)
				require.IsType(t, &definitionv1.Service{}, c)

				service := c.(*definitionv1.Service)

				require.Equal(t, &definitionv1.ContainerBuild{
					DockerBuild: &definitionv1.DockerBuild{
						DockerFile: &definitionv1.DockerFile{
							Location: &definitionv1.Location{
								GitRepository: &definitionv1.GitRepository{
									URL:    fmt.Sprintf("file://%v", remote1Dir),
									Commit: &locationCommitStr,
								},
							},
							Path: definitionv1.DockerBuildDefaultPath,
						},
						BuildContext: &definitionv1.DockerBuildContext{
							Location: &definitionv1.Location{
								GitRepository: &definitionv1.GitRepository{
									URL:    fmt.Sprintf("file://%v", remote1Dir),
									Commit: &locationCommitStr,
								},
							},
							Path: definitionv1.DockerBuildDefaultPath,
						},
					},
				}, service.Container.Build)
			},
		},
	}

	for _, test := range tests {
		setup(test.ServiceBytes)
		t.Run(test.Description, test.Test)
	}
}

func cleanReferenceResolverWorkDir(t *testing.T) {
	err := os.RemoveAll(workDir)
	if err != nil {
		t.Fatal("unable to clean up work directory")
	}
}

func shouldResolveToService(
	r ComponentResolver,
	ref *definitionv1.Reference,
	expected *definitionv1.Service,
) error {
	return shouldResolveToServiceCtx(r, &git.FileReference{}, ref, expected)
}

func shouldResolveToServiceCtx(
	r ComponentResolver,
	ctx *git.FileReference,
	ref *definitionv1.Reference,
	expected *definitionv1.Service,
) error {
	c, err := resolveReference(r, systemID, tree.RootPath(), ctx, ref, true)
	if err != nil {
		return err
	}

	switch typed := c.(type) {
	case *definitionv1.Service:
		if !reflect.DeepEqual(typed, expected) {
			return fmt.Errorf("got invalid contents when resolving git branch reference")
		}

	default:
		return fmt.Errorf("got invalid contents when resolving git commit reference (expected service but got something else)")
	}

	return nil
}

func shouldFailToResolve(
	r ComponentResolver,
	ref *definitionv1.Reference,
) error {
	return shouldFailToResolveCtx(r, &git.FileReference{}, ref)
}

func shouldFailToResolveCtx(
	r ComponentResolver,
	ctx *git.FileReference,
	ref *definitionv1.Reference,
) error {
	_, err := resolveReference(r, systemID, tree.RootPath(), ctx, ref, false)
	return err
}

func resolveReference(
	r ComponentResolver,
	id v1.SystemID,
	p tree.Path,
	ctx *git.FileReference,
	ref *definitionv1.Reference,
	shouldSucceed bool,
) (component.Interface, error) {
	rr, err := r.ResolveReference(id, p, ctx, ref, DepthInfinite)
	if err != nil {
		if !shouldSucceed {
			return nil, nil
		}
		return nil, fmt.Errorf("did not expect error resolving reference but got: %v", err)
	}

	if !shouldSucceed {
		return nil, fmt.Errorf("expected referece resolution to return error but got nil")
	}
	return rr.Component, nil
}
