package app

import (
	"github.com/robfig/cron"
	"github.com/sirupsen/logrus"
)

func (a *App) StartCron() {

	c := cron.New()

	logrus.Infof("services count - %v", len(a.config.Services))

	for folder, service := range a.config.Services {

		err := c.AddFunc(service.Spec, func() {
			a.Invoke(folder, service)
		})
		if err != nil {
			logrus.Errorf("failed to setup backup task for service - %s, error: %v", folder, err)
			continue
		}
	}
	c.Start()

	logrus.Info("application started")
	select {}
}
