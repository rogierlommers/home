package hue_exporter

import (
	hue "github.com/collinux/gohue"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rogierlommers/home/config"
	"github.com/sirupsen/logrus"
)

func NewHue(router *gin.Engine, cfg config.AppConfig) {

	// logrus.Infof("using IP address: %s, token: %s", cfg.HueIPAddress, cfg.HueToken)

	bridge, err := hue.NewBridge(cfg.HueIPAddress)
	if err != nil {
		logrus.Errorf("newBridge error: %s", err)
		return
	}
	bridge.Login(cfg.HueToken)

	prometheus.MustRegister(NewHueCollector("", bridge))
}
