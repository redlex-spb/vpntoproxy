package main

import (
	"github.com/sirupsen/logrus"
	"vpntoproxy/internal/config"
	"vpntoproxy/internal/log"
	"vpntoproxy/internal/server"
)

func main() {
	conf := config.Get()

	log.SetDefaultSettings()

	hs := server.New(conf.Server.Port)
	go hs.Run()
	logrus.Infof("Successfully started server on %d", conf.Server.Port)

	//ui.Create(conf.Basic.Debug)

	defer func() {
		if r := recover(); r != nil {
			logrus.Println("Recovered in f", r)
		}
	}()

	hs.GracefulShutdown()
}
