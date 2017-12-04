package test

import (
	"github.com/mlab-lattice/system/pkg/definition/block"
	jsonutil "github.com/mlab-lattice/system/pkg/util/json"
)

func MockGitRepository() *block.GitRepository {
	return MockCommitGitRepository()
}

func MockGitRepositoryExpectedJSON() []byte {
	return MockCommitGitRepositoryExpectedJSON()
}

func MockCommitGitRepository() *block.GitRepository {
	commit := "0d8934873e7191349a76f8e8b90c417e6b63f65f"
	return &block.GitRepository{
		URL:    "github.com/foo/bar",
		Commit: &commit,
	}
}

func MockCommitGitRepositoryExpectedJSON() []byte {
	return GenerateGitRepositoryExpectedJSON(
		[]byte(`"github.com/foo/bar"`),
		nil,
		[]byte(`"0d8934873e7191349a76f8e8b90c417e6b63f65f"`),
	)
}

func MockTagGitRepository() *block.GitRepository {
	tag := "v1.0.0"
	return &block.GitRepository{
		URL: "github.com/foo/bar",
		Tag: &tag,
	}
}

func MockTagGitRepositoryExpectedJSON() []byte {
	return GenerateGitRepositoryExpectedJSON(
		[]byte(`"github.com/foo/bar"`),
		[]byte(`"v1.0.0"`),
		nil,
	)
}

func GenerateGitRepositoryExpectedJSON(url, tag, commit []byte) []byte {
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
