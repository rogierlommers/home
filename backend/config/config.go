package cfg

import (
	"os"
	"strings"
)

type Config struct {
	StaticDir  string
	Mode       string
	GreedyFile string

	BackendVersion string // injected at build time
}

var Settings Config

func ReadConfig() {
	Settings.StaticDir = os.Getenv("DIST_DIRECTORY")
	Settings.Mode = strings.ToUpper(os.Getenv("MODE"))
	Settings.GreedyFile = os.Getenv("GREEDY_FILE")
}
