package app

import (
	"backuper/config"
	"backuper/pkg"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

type App struct {
	config  *config.Config
	tar     *pkg.Tar
	storage *pkg.Storage
}

var (
	mu        sync.Mutex
	IsRunning bool
)

func New(
	config *config.Config,
	tar *pkg.Tar,
	storage *pkg.Storage,
) *App {

	return &App{
		config:  config,
		tar:     tar,
		storage: storage,
	}
}

func (a *App) Invoke(folder string, service config.ConfigService) {

	mu.Lock()

	if IsRunning {
		return
	}

	IsRunning = true

	mu.Unlock()

	logrus.Infof("start %s backup creating", folder)

	currentDateTime := time.Now().Format("2006-01-02 15-04")
	fileName := fmt.Sprintf("%s_%s", strings.ReplaceAll(currentDateTime, " ", "_"), service.File_Name)

	err := a.tar.Archivate(service.Target_Folder, fileName, service.Excluded_Dirs)
	if err != nil {
		logrus.Errorf("failed to create tar archive: %v", err)
		return
	}

	folderExists, err := a.storage.CheckFolderExist(folder)
	if err != nil {
		logrus.Errorf("failed to check folder exists in storage: %v", err)
		return
	}

	if !folderExists {
		err := a.storage.CreateFolder(folder)
		if err != nil {
			logrus.Errorf("failed to create folder in storage: %v", err)
			return
		}
	}

	logrus.Infof("start %s backup uploading to storage", folder)
	err = a.storage.Upload(fmt.Sprintf("%s/%s", folder, fileName), fmt.Sprintf("storage/%s", fileName))
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
