package resolver

import (
	"fmt"
	"testing"

	"encoding/json"
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

func TestReferenceResolver(t *testing.T) {
	testGitReferenceResolve(t)
}

func testGitReferenceResolve(t *testing.T) {
	testCommitGitReferenceResolve(t)
	testBranchGitReferenceResolve(t)
	//testTagGitReferenceResolve(t)
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

	r, err := NewReferenceResolver(workDir)
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

				c, err := r.ResolveReference(nil, ref)
				if err != nil {
					return err
				}

				switch typed := c.(type) {
				case *defintionv1.Service:
					if !reflect.DeepEqual(typed, &service1) {
						return fmt.Errorf("got invalid contents when resolving git commit reference")
					}

				default:
					return fmt.Errorf("got invalid contents when resolving git commit reference (expected service but got something else)")
				}

				return nil
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

				_, err := r.ResolveReference(nil, ref)
				if err == nil {
					return fmt.Errorf("expected error retrieving invalid git commit file but got nil")
				}

				return nil
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

				_, err := r.ResolveReference(nil, ref)
				if err == nil {
					return fmt.Errorf("expected error retrieving invalid git commit hash but got nil")
				}

				return nil
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

	r, err := NewReferenceResolver(workDir)
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

				c, err := r.ResolveReference(nil, ref)
				if err != nil {
					return err
				}

				switch typed := c.(type) {
				case *defintionv1.Service:
					if !reflect.DeepEqual(typed, &service1) {
						return fmt.Errorf("got invalid contents when resolving git branch reference")
					}

				default:
					return fmt.Errorf("got invalid contents when resolving git commit reference (expected service but got something else)")
				}

				return nil
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

				c, err := r.ResolveReference(nil, ref)
				if err != nil {
					return err
				}

				switch typed := c.(type) {
				case *defintionv1.Service:
					if !reflect.DeepEqual(typed, &service2) {
						return fmt.Errorf("got invalid contents when resolving git branch reference")
					}

				default:
					return fmt.Errorf("got invalid contents when resolving git commit reference (expected service but got something else)")
				}

				return nil
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

				c, err := r.ResolveReference(nil, ref)
				if err != nil {
					return err
				}

				switch typed := c.(type) {
				case *defintionv1.Service:
					if !reflect.DeepEqual(typed, &service3) {
						return fmt.Errorf("got invalid contents when resolving git branch reference")
					}

				default:
					return fmt.Errorf("got invalid contents when resolving git commit reference (expected service but got something else)")
				}

				return nil
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

				_, err := r.ResolveReference(nil, ref)
				if err == nil {
					return fmt.Errorf("expected error retrieving invalid git commit file but got nil")
				}

				return nil
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

				_, err := r.ResolveReference(nil, ref)
				if err == nil {
					return fmt.Errorf("expected error retrieving invalid git branch but got nil")
				}

				return nil
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
