package mock

import (
	"github.com/mlab-lattice/lattice/pkg/definition/block"
	jsonutil "github.com/mlab-lattice/lattice/pkg/util/json"
)

func GitRepository() *block.GitRepository {
	return CommitGitRepository()
}

func GitRepositoryExpectedJSON() []byte {
	return CommitGitRepositoryExpectedJSON()
}

func CommitGitRepository() *block.GitRepository {
	commit := "0d8934873e7191349a76f8e8b90c417e6b63f65f"
	return &block.GitRepository{
		URL:    "github.com/foo/bar",
		Commit: &commit,
	}
}

func CommitGitRepositoryExpectedJSON() []byte {
	return GenerateGitRepositoryExpectedJSON(
		[]byte(`"github.com/foo/bar"`),
		nil,
		[]byte(`"0d8934873e7191349a76f8e8b90c417e6b63f65f"`),
	)
}

func TagGitRepository() *block.GitRepository {
	tag := "v1.0.0"
	return &block.GitRepository{
		URL: "github.com/foo/bar",
		Tag: &tag,
	}
}

func TagGitRepositoryExpectedJSON() []byte {
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
