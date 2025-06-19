package main

import (
	"backuper/app"
	"backuper/config"
	"backuper/pkg"

	"github.com/sirupsen/logrus"
)

func main() {

	config := config.New()
	err := config.LoadConfig()
	if err != nil {
		logrus.Errorf("config load error: %v", err)
		return
	}

	tar := pkg.NewTar()
	storage, err := pkg.NewStorage(
		config.S3.Region,
		config.S3.Access_Key,
		config.S3.Secret_Key,
		config.S3.Endpoint,
		config.S3.Bucket,
	)
	if err != nil {
		logrus.Errorf("s3 storage init error: %v", err)
		return
	}

	a := app.New(
		config, tar, storage,
	)

	a.StartCron()
}
