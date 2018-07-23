package resolver

import (
	"fmt"
	"testing"

	"encoding/json"
	"github.com/mlab-lattice/lattice/pkg/definition/component"
	"github.com/mlab-lattice/lattice/pkg/definition/tree"
	defintionv1 "github.com/mlab-lattice/lattice/pkg/definition/v1"
	"github.com/mlab-lattice/lattice/pkg/util/git"
	"os"
	"reflect"
)

const workDir = "/tmp/lattice/test/pkg/definition/resolver/component_resolver"

var (
	remote1Dir = fmt.Sprintf("%v/remote1", workDir)

	service1 = defintionv1.Service{
		Container: defintionv1.Container{
			Exec: &defintionv1.ContainerExec{
				Command: []string{"foo"},
			},
		},
	}
	service2 = defintionv1.Service{
		Container: defintionv1.Container{
			Exec: &defintionv1.ContainerExec{
				Command: []string{"bar"},
			},
		},
	}
	service3 = defintionv1.Service{
		Container: defintionv1.Container{
			Exec: &defintionv1.ContainerExec{
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

	r, err := NewReferenceResolver(workDir, NewMemoryTemplateStore())
	if err != nil {
		t.Fatal(err)
	}

	commitStr := commit.String()
	ctx := &defintionv1.GitRepositoryReference{
		File: "foo",
		GitRepository: &defintionv1.GitRepository{
			URL:    fmt.Sprintf("file://%v", remote1Dir),
			Commit: &commitStr,
		},
	}

	ref := &defintionv1.Reference{
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
	testTagGitReferenceResolve(t)
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

	servicePath := "service.json"
	commit, err := git.WriteAndCommitFile(remote1Dir, servicePath, serviceBytes, 0700, "my commit")
	if err != nil {
		t.Fatal(err)
	}

	r, err := NewReferenceResolver(workDir, NewMemoryTemplateStore())
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		Description string
		Test        func() error
	}{
		{
			Description: "valid commit",
			Test: func() error {
				commitStr := commit.String()
				ref := &defintionv1.Reference{
					GitRepository: &defintionv1.GitRepositoryReference{
						GitRepository: &defintionv1.GitRepository{
							URL:    fmt.Sprintf("file://%v", remote1Dir),
							Commit: &commitStr,
						},
						File: servicePath,
					},
				}

				return shouldResolveToService(r, ref, &service1)
			},
		},
		{
			Description: "invalid file",
			Test: func() error {
				commitStr := commit.String()
				ref := &defintionv1.Reference{
					GitRepository: &defintionv1.GitRepositoryReference{
						GitRepository: &defintionv1.GitRepository{
							URL:    fmt.Sprintf("file://%v", remote1Dir),
							Commit: &commitStr,
						},
						File: "foo",
					},
				}

				return shouldFailToResolve(r, ref)
			},
		},
		{
			Description: "invalid git commit",
			Test: func() error {
				commitStr := "0123456789012345678901234567890123456789"
				ref := &defintionv1.Reference{
					GitRepository: &defintionv1.GitRepositoryReference{
						GitRepository: &defintionv1.GitRepository{
							URL:    fmt.Sprintf("file://%v", remote1Dir),
							Commit: &commitStr,
						},
						File: "service.json",
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

	r, err := NewReferenceResolver(workDir, NewMemoryTemplateStore())
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
				ref := &defintionv1.Reference{
					GitRepository: &defintionv1.GitRepositoryReference{
						GitRepository: &defintionv1.GitRepository{
							URL:    fmt.Sprintf("file://%v", remote1Dir),
							Branch: &branchName,
						},
						File: servicePath,
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

				ref := &defintionv1.Reference{
					GitRepository: &defintionv1.GitRepositoryReference{
						GitRepository: &defintionv1.GitRepository{
							URL:    fmt.Sprintf("file://%v", remote1Dir),
							Branch: &branchName,
						},
						File: servicePath,
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

				ref := &defintionv1.Reference{
					GitRepository: &defintionv1.GitRepositoryReference{
						GitRepository: &defintionv1.GitRepository{
							URL:    fmt.Sprintf("file://%v", remote1Dir),
							Branch: &branchName,
						},
						File: servicePath,
					},
				}

				return shouldResolveToService(r, ref, &service3)
			},
		},
		{
			Description: "invalid file",
			Test: func() error {
				branchName := "bar"
				ref := &defintionv1.Reference{
					GitRepository: &defintionv1.GitRepositoryReference{
						GitRepository: &defintionv1.GitRepository{
							URL:    fmt.Sprintf("file://%v", remote1Dir),
							Branch: &branchName,
						},
						File: "foo",
					},
				}

				return shouldFailToResolve(r, ref)
			},
		},
		{
			Description: "invalid branch",
			Test: func() error {
				branch := "invalid"
				ref := &defintionv1.Reference{
					GitRepository: &defintionv1.GitRepositoryReference{
						GitRepository: &defintionv1.GitRepository{
							URL:    fmt.Sprintf("file://%v", remote1Dir),
							Branch: &branch,
						},
						File: "service.json",
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

func testTagGitReferenceResolve(t *testing.T) {
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

	r, err := NewReferenceResolver(workDir, NewMemoryTemplateStore())
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

				ref := &defintionv1.Reference{
					GitRepository: &defintionv1.GitRepositoryReference{
						GitRepository: &defintionv1.GitRepository{
							URL: fmt.Sprintf("file://%v", remote1Dir),
							Tag: &tagName,
						},
						File: servicePath,
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
				ref := &defintionv1.Reference{
					GitRepository: &defintionv1.GitRepositoryReference{
						GitRepository: &defintionv1.GitRepository{
							URL: fmt.Sprintf("file://%v", remote1Dir),
							Tag: &patchSemverTag,
						},
						File: servicePath,
					},
				}

				return shouldResolveToService(r, ref, &service1)
			},
		},
		{
			Description: "strict semver minor should initially resolve",
			Test: func() error {
				minorSemverTag := "1.x"
				ref := &defintionv1.Reference{
					GitRepository: &defintionv1.GitRepositoryReference{
						GitRepository: &defintionv1.GitRepository{
							URL: fmt.Sprintf("file://%v", remote1Dir),
							Tag: &minorSemverTag,
						},
						File: servicePath,
					},
				}

				return shouldResolveToService(r, ref, &service1)
			},
		},
		{
			Description: "strict semver invalid major should not initially resolve",
			Test: func() error {
				invalidSemverTag := "2.x"
				ref := &defintionv1.Reference{
					GitRepository: &defintionv1.GitRepositoryReference{
						GitRepository: &defintionv1.GitRepository{
							URL: fmt.Sprintf("file://%v", remote1Dir),
							Tag: &invalidSemverTag,
						},
						File: servicePath,
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
				ref := &defintionv1.Reference{
					GitRepository: &defintionv1.GitRepositoryReference{
						GitRepository: &defintionv1.GitRepository{
							URL: fmt.Sprintf("file://%v", remote1Dir),
							Tag: &patchTag,
						},
						File: servicePath,
					},
				}

				return shouldResolveToService(r, ref, &service2)
			},
		},
		{
			Description: "strict semver minor resolve should update with patch update",
			Test: func() error {
				minorTag := "1.x"
				ref := &defintionv1.Reference{
					GitRepository: &defintionv1.GitRepositoryReference{
						GitRepository: &defintionv1.GitRepository{
							URL: fmt.Sprintf("file://%v", remote1Dir),
							Tag: &minorTag,
						},
						File: servicePath,
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
				ref := &defintionv1.Reference{
					GitRepository: &defintionv1.GitRepositoryReference{
						GitRepository: &defintionv1.GitRepository{
							URL: fmt.Sprintf("file://%v", remote1Dir),
							Tag: &patchSemverTag,
						},
						File: servicePath,
					},
				}

				return shouldResolveToService(r, ref, &service2)
			},
		},
		{
			Description: "strict semver minor resolve should update with minor update",
			Test: func() error {
				minorSemverTag := "1.x"
				ref := &defintionv1.Reference{
					GitRepository: &defintionv1.GitRepositoryReference{
						GitRepository: &defintionv1.GitRepository{
							URL: fmt.Sprintf("file://%v", remote1Dir),
							Tag: &minorSemverTag,
						},
						File: servicePath,
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
				ref := &defintionv1.Reference{
					GitRepository: &defintionv1.GitRepositoryReference{
						GitRepository: &defintionv1.GitRepository{
							URL: fmt.Sprintf("file://%v", remote1Dir),
							Tag: &patchSemverTag,
						},
						File: servicePath,
					},
				}

				return shouldResolveToService(r, ref, &service2)
			},
		},
		{
			Description: "strict semver minor resolve should not update with major update",
			Test: func() error {
				minorSemverTag := "1.x"
				ref := &defintionv1.Reference{
					GitRepository: &defintionv1.GitRepositoryReference{
						GitRepository: &defintionv1.GitRepository{
							URL: fmt.Sprintf("file://%v", remote1Dir),
							Tag: &minorSemverTag,
						},
						File: servicePath,
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

func cleanReferenceResolverWorkDir(t *testing.T) {
	err := os.RemoveAll(workDir)
	if err != nil {
		t.Fatal("unable to clean up work directory")
	}
}

func shouldResolveToService(
	r ReferenceResolver,
	ref *defintionv1.Reference,
	expected *defintionv1.Service,
) error {
	return shouldResolveToServiceCtx(r, &defintionv1.GitRepositoryReference{GitRepository: &defintionv1.GitRepository{}}, ref, expected)
}

func shouldResolveToServiceCtx(
	r ReferenceResolver,
	ctx *defintionv1.GitRepositoryReference,
	ref *defintionv1.Reference,
	expected *defintionv1.Service,
) error {
	c, err := resolveReference(r, tree.RootNodePath(), ctx, ref, true)
	if err != nil {
		return err
	}

	switch typed := c.(type) {
	case *defintionv1.Service:
		if !reflect.DeepEqual(typed, expected) {
			return fmt.Errorf("got invalid contents when resolving git branch reference")
		}

	default:
		return fmt.Errorf("got invalid contents when resolving git commit reference (expected service but got something else)")
	}

	return nil
}

func shouldFailToResolve(
	r ReferenceResolver,
	ref *defintionv1.Reference,
) error {
	return shouldFailToResolveCtx(r, &defintionv1.GitRepositoryReference{GitRepository: &defintionv1.GitRepository{}}, ref)
}

func shouldFailToResolveCtx(
	r ReferenceResolver,
	ctx *defintionv1.GitRepositoryReference,
	ref *defintionv1.Reference,
) error {
	_, err := resolveReference(r, tree.RootNodePath(), ctx, ref, false)
	return err
}

func resolveReference(
	r ReferenceResolver,
	p tree.NodePath,
	ctx *defintionv1.GitRepositoryReference,
	ref *defintionv1.Reference,
	shouldSucceed bool,
) (component.Interface, error) {
	c, _, err := r.ResolveReference(p, ctx, ref)
	if err != nil {
		if !shouldSucceed {
			return nil, nil
		}
		return nil, fmt.Errorf("did not expect error resolving reference but got: %v", err)
	}

	if !shouldSucceed {
		return nil, fmt.Errorf("expected referece resolution to return error but got nil")
	}
	return c, nil
}
