package drivers

import "fmt"

type (
	FTP struct {
	}

	FTPConfig struct {
		Host     string
		User     string
		Password string
		Port     int
		Folder   string
	}
)

func NewFTP() DriverInterface {

	return &FTP{}
}

func (d *FTP) Init(cfgData any) error {

	cfg, ok := cfgData.(FTPConfig)
	if !ok {
		return fmt.Errorf("config type must be - `S3Config`")
	}

	fmt.Println("cfg", cfg)

	return nil
}

func (d *FTP) Ping() error {

	return nil
}

func (d *FTP) CheckFolderExist(folderName string) (bool, error) {

	return true, nil
}

func (d *FTP) CreateFolder(folderName string) error {

	return nil
}

func (d *FTP) Upload(folderName, filePath string) error {

	return nil
}
