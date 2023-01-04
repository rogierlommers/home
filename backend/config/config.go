package cfg

import (
	"os"
	"strings"
)

type Config struct {
	StaticDir      string
	BackendVersion string
	Mode           string
}

var Settings Config

func ReadConfig() {
	Settings.StaticDir = os.Getenv("DIST_DIRECTORY")
	Settings.Mode = strings.ToUpper(os.Getenv("MODE"))
	Settings.BackendVersion = "poep"
}
