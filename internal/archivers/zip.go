package archivers

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Zip struct {
}

func NewZip() ArchiverInterface {

	return &Zip{}
}

func (a *Zip) Archivate(sourceDir, targetFile string, excludeDirs []string) (string, error) {

	targetFile = fmt.Sprintf("%s.zip", targetFile)
	targetFilePath := fmt.Sprintf("storage/%s", targetFile)

	file, err := os.Create(targetFilePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	zipWriter := zip.NewWriter(file)
	defer zipWriter.Close()

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

		fileToZip, err := os.Open(path)
		if err != nil {
			return err
		}
		defer fileToZip.Close()

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = relPath
		header.Method = zip.Deflate

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}

		_, err = io.Copy(writer, fileToZip)
		return err
	})
}
