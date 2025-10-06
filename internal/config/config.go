package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

type AppConfig struct {
	HostPort     string
	GreedyFile   string
	Database     string
	UploadTarget string
	Username     string
	Password     string
	CleanUpInDys int
}

func ReadConfig() AppConfig {

	c := AppConfig{
		GreedyFile:   os.Getenv("GREEDY_FILE"),
		Database:     os.Getenv("DATABASE"),
		UploadTarget: os.Getenv("UPLOAD_TARGET"),
		Username:     os.Getenv("USERNAME"),
		Password:     os.Getenv("PASSWORD"),
	}

	// host and port
	if strings.ToLower(os.Getenv("DEV")) == "true" {
		c.HostPort = "127.0.0.1:3000"
		logrus.Info("develoment mode, debug level logging enabled")
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		c.HostPort = ":3000"
		logrus.Info("production mode, error level logging enabled")
		logrus.SetLevel(logrus.ErrorLevel)
	}

	// cleanup days
	if days := os.Getenv("CLEANUP_DAYS"); days != "" {
		var err error
		_, err = fmt.Sscanf(days, "%d", &c.CleanUpInDys)
		if err != nil {
			logrus.Errorf("invalid CLEANUP_DAYS value, defaulting to 30 days: %v", err)
			c.CleanUpInDys = 30
		}
	} else {
		logrus.Error("invalid CLEANUP_DAYS value, defaulting to 30 days")
		c.CleanUpInDys = 30
	}

	return c
}
