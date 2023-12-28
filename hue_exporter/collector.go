package hue_exporter

import (
	"fmt"
	"regexp"

	hue "github.com/collinux/gohue"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

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
				"reachable",
			},
		),
	}

	return c
}

func (c hueCollector) Describe(ch chan<- *prometheus.Desc) {
	c.lightState.Describe(ch)
}

func (c hueCollector) Collect(ch chan<- prometheus.Metric) {

	lights, err := c.bridge.GetAllLights()
	if err != nil {
		logrus.Errorf("Failed to update lights: %v", err)
		return
	}

	nameRe := regexp.MustCompile("[^a-zA-Z0-9_]")

	for _, light := range lights {

		reachableValue := fmt.Sprintf("%v", light.State.Reachable)
		name := nameRe.ReplaceAllString(light.Name, "_")

		// logrus.Infof("Name: %s / turned on: %v, reachable: %v", light.Name, light.State.On, reachableValue)
		if !light.State.Reachable {
			c.lightState.With(prometheus.Labels{"name": name, "reachable": reachableValue}).Set(-1.0)
		} else if light.State.On {
			c.lightState.With(prometheus.Labels{"name": name, "reachable": reachableValue}).Set(1.0)
		} else {
			c.lightState.With(prometheus.Labels{"name": name, "reachable": reachableValue}).Set(0.0)
		}

	}

	c.lightState.Collect(ch)
}
