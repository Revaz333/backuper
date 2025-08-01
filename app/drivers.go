package app

import (
	"backuper/internal/drivers"
	"fmt"

	"github.com/sirupsen/logrus"
)

// func (a *App) Invoke(folder string, service config.ConfigService) {

// 	mu.Lock()

// 	if IsRunning {
// 		return
// 	}

// 	IsRunning = true

// 	mu.Unlock()

// 	logrus.Infof("start %s backup creating", folder)

// 	currentDateTime := time.Now().Format("2006-01-02 15-04")
// 	fileName := fmt.Sprintf("%s_%s", strings.ReplaceAll(currentDateTime, " ", "_"), service.File_Name)

// 	err := a.tar.Archivate(service.Target_Folder, fileName, service.Excluded_Dirs)
// 	if err != nil {
// 		logrus.Errorf("failed to create tar archive: %v", err)
// 		return
// 	}

// 	folderExists, err := a.storage.CheckFolderExist(folder)
// 	if err != nil {
// 		logrus.Errorf("failed to check folder exists in storage: %v", err)
// 		return
// 	}

// 	if !folderExists {
// 		err := a.storage.CreateFolder(folder)
// 		if err != nil {
// 			logrus.Errorf("failed to create folder in storage: %v", err)
// 			return
// 		}
// 	}

// 	logrus.Infof("start %s backup uploading to storage", folder)
// 	err = a.storage.Upload(fmt.Sprintf("%s/%s", folder, fileName), fmt.Sprintf("storage/%s", fileName))
// 	if err != nil {
// 		logrus.Errorf("failed to upload new object to folder in storage: %v", err)
// 		return
// 	}
// 	logrus.Infof("%s backup uploadgin done", folder)

// 	err = os.Remove(fmt.Sprintf("storage/%s", fileName))
// 	if err != nil {
// 		logrus.Errorf("failed to remove backup tmp file from local storage: %v", err)
// 		return
// 	}

// 	logrus.Infof("%s backup creating done", folder)

// 	mu.Lock()

// 	IsRunning = false

// 	mu.Unlock()
// }

func (a *App) ValidateDrivers() error {

	logrus.Info("porcess drivers validating...")

	if len(a.config.Drivers.S3) != 0 {
		for profile, config := range a.config.Drivers.S3 {

			driver := drivers.NewS3()
			err := driver.Init(drivers.S3Config{
				Endpoint:  config.Endpoint,
				Bucket:    config.Bucket,
				AccessKey: config.Access_Key,
				SecretKey: config.Secret_Key,
				Region:    config.Region,
			})
			if err != nil {
				return fmt.Errorf("failed to init s3 connection with profile - `%s`, got error: %v", profile, err)
			}

			err = driver.Ping()
			if err != nil {
				return fmt.Errorf("failed to connect to s3 bucket with profile - `%s`, got error: %v", profile, err)
			}
		}
	}

	return nil
}

func (a *App) GetDriverConfig(driver, profile string) (any, error) {

	switch driver {
	case DriverS3:
		if _, found := a.config.Drivers.S3[profile]; !found {
			return nil, fmt.Errorf("profile - `%s` for `s3` driver not found", profile)
		}

		config := a.config.Drivers.S3[profile]

		return drivers.S3Config{
			Endpoint:  config.Endpoint,
			Bucket:    config.Bucket,
			AccessKey: config.Access_Key,
			SecretKey: config.Secret_Key,
			Region:    config.Region,
		}, nil

	case DriverFTP:
		if _, found := a.config.Drivers.S3[profile]; !found {
			return nil, fmt.Errorf("profile - `%s` for `ftp` driver not found", profile)
		}

		config := a.config.Drivers.FTP[profile]

		return drivers.FTPConfig{
			Host:     config.Host,
			User:     config.User,
			Password: config.Passwrod,
			Port:     config.Port,
			Folder:   config.Folder,
		}, nil

	default:
		return nil, fmt.Errorf("driver - `%s` not found", driver)
	}
}
