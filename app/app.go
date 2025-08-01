package app

import (
	"backuper/config"
	"backuper/internal/archivers"
	"backuper/internal/drivers"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/robfig/cron"
	"github.com/sirupsen/logrus"
)

type App struct {
	config *config.Config
}

var (
	mu        sync.Mutex
	IsRunning bool
)

const (
	DriverS3  = "s3"
	DriverFTP = "ftp"

	ArchiverZip = "zip"
	ArchiverTar = "tar"
)

func New(
	config *config.Config,
) *App {

	return &App{
		config: config,
	}
}

func (a *App) StartApp() {

	logrus.Info("starting application...")

	err := a.ValidateDrivers()
	if err != nil {
		logrus.Errorf("driver error: %v", err)
		return
	}

	c := cron.New()

	logrus.Infof("services count - %v", len(a.config.Services))

	for service, serviceCfg := range a.config.Services {

		archivers := archivers.NewArchiver().LoadAllowedDrivers()

		archiver, allowed := archivers[serviceCfg.Archiver]
		if !allowed {
			logrus.Errorf("archiver - `%s` not allowed", serviceCfg.Archiver)
			continue
		}

		drivers := drivers.NewDriver().LoadAllowedDrivers()

		driver, allowed := drivers[serviceCfg.Driver.Type]
		if !allowed {
			logrus.Errorf("driver - `%s` not allowed", serviceCfg.Driver.Type)
			continue
		}

		drvierCfg, err := a.GetDriverConfig(serviceCfg.Driver.Type, serviceCfg.Driver.Profile)
		if err != nil {
			logrus.Errorf("failed to load `%s` drvier config by profile - `%s`: %v", serviceCfg.Driver.Type, serviceCfg.Driver.Profile, err)
			return
		}

		drv := driver()
		err = drv.Init(drvierCfg)
		if err != nil {
			logrus.Errorf("an error occured while init driver - `%s|%s`: %v", serviceCfg.Driver.Type, serviceCfg.Driver.Profile, err)
		}

		err = c.AddFunc(serviceCfg.Spec, func() {

			mu.Lock()
			if IsRunning {
				return
			}

			IsRunning = true
			mu.Unlock()

			a.Backup(archiver(), drv, service, serviceCfg)
			err = drv.Cleanup(service, serviceCfg.Max_Images)
			if err != nil {
				logrus.Errorf("cleanup action failed: %v", err)

				mu.Lock()
				IsRunning = false
				mu.Unlock()

				return
			}
			logrus.Info("cleanup action success.")

			mu.Lock()
			IsRunning = false
			mu.Unlock()
		})
		if err != nil {
			logrus.Errorf("failed to setup backup task for service - `%s`: %v", service, err)
			continue
		}
	}

	c.Start()

	select {}
}

func (a *App) Backup(
	archiver archivers.ArchiverInterface,
	driver drivers.DriverInterface,
	folder string,
	service config.ConfigService,
) {

	logrus.Infof("start %s backup creating", folder)

	currentDateTime := time.Now().Format("2006-01-02 15-04")
	fileName := fmt.Sprintf("%s_%s", strings.ReplaceAll(currentDateTime, " ", "_"), service.File_Name)

	fileName, err := archiver.Archivate(service.Target_Folder, fileName, service.Excluded_Dirs)
	if err != nil {
		logrus.Errorf("failed to create archive: %v", err)
		return
	}

	/* soon
	folderExists, err := driver.CheckFolderExist(folder)
	if err != nil {
		logrus.Errorf("failed to check folder exists in storage: %v", err)
		return
	}

	if !folderExists {
		err := driver.CreateFolder(folder)
		if err != nil {
			logrus.Errorf("failed to create folder in storage: %v", err)
			return
		}
	}
	*/

	logrus.Infof("start %s backup uploading to storage", folder)

	err = driver.Upload(fmt.Sprintf("%s/%s", folder, fileName), fmt.Sprintf("storage/%s", fileName))
	if err != nil {
		logrus.Errorf("failed to upload new object to folder in storage: %v", err)
		return
	}

	logrus.Infof("%s backup uploadgin done", folder)

	err = os.Remove(fmt.Sprintf("storage/%s", fileName))
	if err != nil {
		logrus.Errorf("failed to remove backup tmp file from local storage: %v", err)
		return
	}

	logrus.Infof("%s backup creating done", folder)
}
