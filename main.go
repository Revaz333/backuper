package main

import (
	"backuper/app"
	"backuper/config"

	"github.com/sirupsen/logrus"
)

func main() {

	config := config.New()
	err := config.LoadConfig()
	if err != nil {
		logrus.Errorf("config load error: %v", err)
		return
	}

	// tar := pkg.NewTar()
	// storage, err := pkg.NewStorage(
	// 	config.Drivers.S3.Region,
	// 	config.Drivers.S3.Access_Key,
	// 	config.Drivers.S3.Secret_Key,
	// 	config.Drivers.S3.Endpoint,
	// 	config.Drivers.S3.Bucket,
	// )
	// if err != nil {
	// 	logrus.Errorf("s3 storage init error: %v", err)
	// 	return
	// }

	a := app.New(config)

	a.StartApp()
}
