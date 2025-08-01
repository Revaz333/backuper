package app

import (
	"backuper/config"
	"backuper/internal/archivers"
	"backuper/internal/drivers"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/robfig/cron"
	"github.com/sirupsen/logrus"
)

type App struct {
	config *config.Config
	// tar    *pkg.Tar
	// storage *pkg.Storage
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

		fmt.Println("ddd", service, serviceCfg)
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

		err = c.AddFunc(serviceCfg.Spec, func() {

			a.Backup(archiver(), driver(), service, serviceCfg)
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

	mu.Lock()

	if IsRunning {
		return
	}

	IsRunning = true

	mu.Unlock()

	logrus.Infof("start %s backup creating", folder)

	currentDateTime := time.Now().Format("2006-01-02 15-04")
	fileName := fmt.Sprintf("%s_%s", strings.ReplaceAll(currentDateTime, " ", "_"), service.File_Name)

	fileName, err := archiver.Archivate(service.Target_Folder, fileName, service.Excluded_Dirs)
	if err != nil {
		logrus.Errorf("failed to create archive: %v", err)
		return
	}

	drvierCfg, err := a.GetDriverConfig(service.Driver.Type, service.Driver.Profile)
	if err != nil {
		logrus.Errorf("failed to load `%s` drvier config by profile - `%s`: %v", service.Driver.Type, service.Driver.Profile, err)
		return
	}

	err = driver.Init(drvierCfg)
	if err != nil {
		logrus.Errorf("an error occured while init driver - `%s|%s`: %v", service.Driver.Type, service.Driver.Profile, err)
	}

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

	err = filepath.Walk(localDir, func(path string, info os.FileInfo, err error) error {

		filesCount++
		return nil
	})

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

	mu.Lock()

	IsRunning = false

	mu.Unlock()
}
