package test

import (
	"reflect"
	"testing"

	"github.com/mlab-lattice/system/pkg/definition/block"
)

func TestGitRepository_Validate(t *testing.T) {
	tag := "v1.0.0"
	commit := "0d8934873e7191349a76f8e8b90c417e6b63f65f"
	Validate(
		t,
		nil,

		// Invalid GitRepo
		[]ValidateTest{
			{
				Description:     "empty",
				DefinitionBlock: &block.GitRepository{},
			},
			{
				Description: "Tag and Commit",
				DefinitionBlock: &block.GitRepository{
					Tag:    &tag,
					Commit: &commit,
				},
			},
		},

		// Valid Builds
		[]ValidateTest{
			{
				Description: "Tag",
				DefinitionBlock: &block.GitRepository{
					Tag: &tag,
				},
			},
			{
				Description: "Commit",
				DefinitionBlock: &block.GitRepository{
					Commit: &commit,
				},
			},
		},
	)
}

func TestGitRepository_JSON(t *testing.T) {
	JSON(
		t,
		reflect.TypeOf(block.GitRepository{}),
		[]JSONTest{
			{
				Description: "MockCommitGitRepository",
				Bytes:       MockCommitGitRepositoryExpectedJSON(),
				ValuePtr:    MockCommitGitRepository(),
			},
			{
				Description: "MockTagGitRepository",
				Bytes:       MockTagGitRepositoryExpectedJSON(),
				ValuePtr:    MockTagGitRepository(),
			},
		},
	)
}
