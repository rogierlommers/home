package cfg

import (
	"os"
	"strings"
)

type Config struct {
	StaticDir      string
	BackendVersion string
	Mode           string
	GreedyFile     string
}

var Settings Config

func ReadConfig() {
	Settings.StaticDir = os.Getenv("DIST_DIRECTORY")
	Settings.Mode = strings.ToUpper(os.Getenv("MODE"))
	Settings.GreedyFile = os.Getenv("GREEDY_FILE")

	// injected at build-time
	Settings.BackendVersion = "poep"
}
