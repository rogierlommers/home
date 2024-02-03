package config

import (
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

type AppConfig struct {
	HostPort       string
	GreedyFile     string
	EnyaqVIN       string
	EnyaqUsername  string
	EnyaqPassword  string
	HueIPAddress   string
	HueToken       string
	HomeWizardHost string
}

func ReadConfig() AppConfig {

	c := AppConfig{
		GreedyFile:     os.Getenv("GREEDY_FILE"),
		EnyaqVIN:       os.Getenv("ENYAQ_VIN"),
		EnyaqUsername:  os.Getenv("ENYAQ_USERNAME"),
		EnyaqPassword:  os.Getenv("ENYAQ_PASSWORD"),
		HueIPAddress:   os.Getenv("HUE_IPADDRESS"),
		HueToken:       os.Getenv("HUE_TOKEN"),
		HomeWizardHost: os.Getenv("HOMEWIZARD_HOST"),
	}

	if strings.ToLower(os.Getenv("DEV")) == "true" {
		c.HostPort = "127.0.0.1:3000"
		logrus.Info("develoment mode")
	} else {
		c.HostPort = ":3000"
	}

	return c
}
