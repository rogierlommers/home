package hue_exporter

import (
	"regexp"

	hue "github.com/collinux/gohue"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rogierlommers/home/config"
	"github.com/sirupsen/logrus"
)

// inspiration: https://grafana.com/grafana/dashboards/13645-philips-hue/

func NewHue(router *gin.Engine, cfg config.AppConfig) {

	bridge, err := hue.NewBridge(cfg.HueIPAddress)
	if err != nil {
		logrus.Errorf("newBridge error: %s", err)
		return
	}
	bridge.Login(cfg.HueToken)

	prometheus.MustRegister(NewHueCollector("", bridge))
}

type hueCollector struct {
	bridge     *hue.Bridge
	lightState *prometheus.GaugeVec
}

func NewHueCollector(namespace string, bridge *hue.Bridge) prometheus.Collector {
	c := hueCollector{
		bridge: bridge,
		lightState: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "hue_lights",
				Name:      "state",
				Help:      "Lights on/off state",
			},
			[]string{
				"name",
				// "reachable",
			},
		),
	}

	return c
}

func (c hueCollector) Describe(ch chan<- *prometheus.Desc) {
	c.lightState.Describe(ch)
}

func (c hueCollector) Collect(ch chan<- prometheus.Metric) {

	// First collect information about all lights
	lights, err := c.bridge.GetAllLights()
	if err != nil {
		logrus.Errorf("Failed to update lights: %v", err)
		return
	}

	nameRe := regexp.MustCompile("[^a-zA-Z0-9_]")

	for _, light := range lights {

		name := nameRe.ReplaceAllString(light.Name, "_")

		if !light.State.Reachable {
			c.lightState.With(prometheus.Labels{"name": name}).Set(-1.0)
		} else if light.State.On {
			c.lightState.With(prometheus.Labels{"name": name}).Set(1.0)
		} else {
			c.lightState.With(prometheus.Labels{"name": name}).Set(0.0)
		}
	}
	c.lightState.Collect(ch)

	// Then collect information about all sensors
	// sensors, err := c.bridge.GetAllSensors()
	// if err != nil {
	// 	logrus.Errorf("Failed to update sensors: %v", err)
	// 	return
	// }

	// for _, sensor := range sensors {

	// 	if sensor.Name == "HomeAway" {
	// 		// logrus.Infof("donderop, want sensor.name == \"HomeAway\"")
	// 		continue
	// 	}

	// 	if sensor.Type == "CLIPGenericStatus" {
	// 		// logrus.Infof("donderop, want sensor.modelID == \"CLIPGenericStatus\"")
	// 		continue
	// 	}

	// 	if sensor.Type == "Daylight" {
	// 		// logrus.Infof("donderop, want sensor.Type == \"Daylight\"")
	// 		continue
	// 	}

	// 	if sensor.Type == "ZLLLightLevel" {
	// 		continue
	// 	}

	// 	// spew.Dump(sensor)
	// 	logrus.Infof("uniqueID: %s, modelID: %s, type: %s, name: %s, battery level: %d", sensor.UniqueID, sensor.ModelID, sensor.Type, sensor.Name, sensor.Config.Battery)
	// }
}
