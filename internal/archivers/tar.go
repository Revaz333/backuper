package archivers

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Tar struct {
}

func NewTar() ArchiverInterface {

	return &Tar{}
}

func (a *Tar) Archivate(sourceDir, targetFile string, excludeDirs []string) (string, error) {

	targetFile = fmt.Sprintf("%s.tar.gz", targetFile)
	targetFilePath := fmt.Sprintf("storage/%s", targetFile)

	file, err := os.Create(targetFilePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	gw := gzip.NewWriter(file)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	return targetFile, filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}
		for _, excl := range excludeDirs {
			if relPath == excl || strings.HasPrefix(relPath, excl+"/") {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		if info.IsDir() {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		header.Name = relPath

		if err := tw.WriteHeader(header); err != nil {
			return err
		}
		_, err = io.Copy(tw, f)
		return err
	})
}
