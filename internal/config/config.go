package config

import (
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

type AppConfig struct {
	HostPort     string
	GreedyFile   string
	UploadTarget string
	Username     string
	Password     string
}

func ReadConfig() AppConfig {

	c := AppConfig{
		GreedyFile:   os.Getenv("GREEDY_FILE"),
		UploadTarget: os.Getenv("UPLOAD_TARGET"),
		Username:     os.Getenv("USERNAME"),
		Password:     os.Getenv("PASSWORD"),
	}

	if strings.ToLower(os.Getenv("DEV")) == "true" {
		c.HostPort = "127.0.0.1:3000"
		logrus.Info("develoment mode, debug level logging enabled")
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		c.HostPort = ":3000"
		logrus.Info("production mode, error level logging enabled")
		logrus.SetLevel(logrus.ErrorLevel)
	}

	return c
}
