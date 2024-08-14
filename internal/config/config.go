package config

import (
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

type AppConfig struct {
	HostPort   string
	GreedyFile string
}

func ReadConfig() AppConfig {

	c := AppConfig{
		GreedyFile: os.Getenv("GREEDY_FILE"),
	}

	if strings.ToLower(os.Getenv("DEV")) == "true" {
		c.HostPort = "127.0.0.1:3000"
		logrus.Info("develoment mode")
	} else {
		c.HostPort = ":3000"
	}

	return c
}
