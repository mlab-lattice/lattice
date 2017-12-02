package test

import (
	"github.com/mlab-lattice/system/pkg/definition/block"
	jsonutil "github.com/mlab-lattice/system/pkg/util/json"
)

func MockGitRepository() *block.GitRepository {
	return MockCommitGitRepository()
}

func MockGitRepositoryExpectedJson() []byte {
	return MockCommitGitRepositoryExpectedJson()
}

func MockCommitGitRepository() *block.GitRepository {
	commit := "0d8934873e7191349a76f8e8b90c417e6b63f65f"
	return &block.GitRepository{
		Url:    "github.com/foo/bar",
		Commit: &commit,
	}
}

func MockCommitGitRepositoryExpectedJson() []byte {
	return GenerateGitRepositoryExpectedJson(
		[]byte(`"github.com/foo/bar"`),
		nil,
		[]byte(`"0d8934873e7191349a76f8e8b90c417e6b63f65f"`),
	)
}

func MockTagGitRepository() *block.GitRepository {
	tag := "v1.0.0"
	return &block.GitRepository{
		Url: "github.com/foo/bar",
		Tag: &tag,
	}
}

func MockTagGitRepositoryExpectedJson() []byte {
	return GenerateGitRepositoryExpectedJson(
		[]byte(`"github.com/foo/bar"`),
		[]byte(`"v1.0.0"`),
		nil,
	)
}

func GenerateGitRepositoryExpectedJson(url, tag, commit []byte) []byte {
	return jsonutil.GenerateObjectBytes([]jsonutil.FieldBytes{
		{
			Name:  "url",
			Bytes: url,
		},
		{
			Name:      "tag",
			Bytes:     tag,
			OmitEmpty: true,
		},
		{
			Name:      "commit",
			Bytes:     commit,
			OmitEmpty: true,
		},
	})
}
