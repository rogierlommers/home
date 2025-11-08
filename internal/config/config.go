package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

type AppConfig struct {
	HostPort                string
	GreedyFile              string
	GreedyCleanupFrequency  int
	GreedyScrapingFrequency int
	Database                string
	UploadTarget            string
	Username                string
	Password                string
	XHomeAPIKey             string
	FileCleanUpInDys        int
}

func ReadConfig() AppConfig {

	c := AppConfig{
		GreedyFile:              os.Getenv("GREEDY_FILE"),
		GreedyCleanupFrequency:  convertToInt(os.Getenv("GREEDY_CLEANUP_FREQUENCY"), 86400), // default 1 day
		GreedyScrapingFrequency: convertToInt(os.Getenv("GREEDY_SCRAPING_FREQUENCY"), 3600), // default 1 hour
		Database:                os.Getenv("DATABASE"),
		UploadTarget:            os.Getenv("UPLOAD_TARGET"),
		Username:                os.Getenv("USERNAME"),
		Password:                os.Getenv("PASSWORD"),
		XHomeAPIKey:             os.Getenv("X_HOME_API_KEY"),
	}

	// host and port
	c.HostPort = ":3000"

	if strings.ToLower(os.Getenv("DEV")) == "true" {
		logrus.Info("develoment mode, debug level logging enabled")
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.Info("production mode, error level logging enabled")
		logrus.SetLevel(logrus.ErrorLevel)
	}

	// cleanup days
	if days := os.Getenv("FILE_CLEANUP_DAYS"); days != "" {
		var err error
		_, err = fmt.Sscanf(days, "%d", &c.FileCleanUpInDys)
		if err != nil {
			logrus.Errorf("invalid FILE_CLEANUP_DAYS value, defaulting to 30 days: %v", err)
			c.FileCleanUpInDys = 30
		}
	} else {
		logrus.Error("invalid FILE_CLEANUP_DAYS value, defaulting to 30 days")
		c.FileCleanUpInDys = 30
	}

	return c
}

func convertToInt(i string, defaultValue int) int {

	var value int
	_, err := fmt.Sscanf(i, "%d", &value)
	if err != nil {
		logrus.Errorf("invalid integer value, defaulting to %d: %v", defaultValue, err)
		return defaultValue
	}

	return value
}
