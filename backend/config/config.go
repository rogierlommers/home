package cfg

import (
	"os"
	"strings"
)

var BackendVersion string

type Config struct {
	StaticDir      string
	Mode           string
	GreedyFile     string
	EnyaqVIN       string
	EnyaqUsername  string
	EnyaqPassword  string
	BackendVersion string // injected at build time
}

var Settings Config

func ReadConfig() {
	Settings.StaticDir = os.Getenv("DIST_DIRECTORY")
	Settings.Mode = strings.ToUpper(os.Getenv("MODE"))
	Settings.GreedyFile = os.Getenv("GREEDY_FILE")
	Settings.BackendVersion = BackendVersion
	Settings.EnyaqVIN = os.Getenv("ENYAQ_VIN")
	Settings.EnyaqUsername = os.Getenv("ENYAQ_USERNAME")
	Settings.EnyaqPassword = os.Getenv("ENYAQ_PASSWORD")
}
