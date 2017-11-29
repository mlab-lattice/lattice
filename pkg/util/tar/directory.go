package tar

import (
	"archive/tar"
	"bytes"
	"io"
	"os"
	"path/filepath"
)

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

		if err = tw.WriteHeader(header); err != nil {
			return err
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
