package tar

import (
	"archive/tar"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ArchiveDirectory takes in the path to a directory and will attempt to return an io.Reader
// containing the directory as a tarball.
func ArchiveDirectory(directoryRoot string) (io.Reader, error) {
	// With a little help from https://golang.org/src/archive/tar/example_test.go
	// and https://medium.com/@skdomino/taring-untaring-files-in-go-6b07cf56bc07
	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)
	defer tw.Close()

	err := filepath.Walk(directoryRoot, func(file string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := tar.FileInfoHeader(fi, fi.Name())
		if err != nil {
			return err
		}

		header.Name = strings.TrimPrefix(strings.Replace(file, directoryRoot, "", -1), string(filepath.Separator))

		if err = tw.WriteHeader(header); err != nil {
			return err
		}

		if !fi.Mode().IsRegular() {
			return nil
		}

		f, err := os.Open(file)
		if err != nil {
			return err
		}
		defer f.Close()

		if _, err := io.Copy(tw, f); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return buf, nil
}
