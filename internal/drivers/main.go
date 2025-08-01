package drivers

import (
	"crypto/sha256"
	"encoding/hex"
	"path/filepath"
)

type (
	DriverInterface interface {
		Init(cfg any) error
		Ping() error
		CheckFolderExist(folderName string) (bool, error)
		CreateFolder(folderName string) error
		Upload(folderName, filePath string) error
		Cleanup(folderName string, maxImages int) error
	}

	Driver struct {
	}
)

func NewDriver() *Driver {

	return &Driver{}
}

func (d *Driver) LoadAllowedDrivers() map[string]func() DriverInterface {

	return map[string]func() DriverInterface{
		"s3":  NewS3,
		"ftp": NewFTP,
	}
}

func calculateSHA256(content []byte) string {

	hash := sha256.Sum256(content)
	return hex.EncodeToString(hash[:])
}

func detectMimeType(path string) string {

	ext := filepath.Ext(path)
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".txt":
		return "text/plain"
	case ".pdf":
		return "application/pdf"
	default:
		return "application/octet-stream"
	}
}
