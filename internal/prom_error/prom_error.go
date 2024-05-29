package prom_error

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

var numberOfErrors = prometheus.NewCounter(prometheus.CounterOpts{
	Name: "home_errors_total",
	Help: "Total number of errors occured in home app",
})

func InitPromError() {
	logrus.Info("initialising prom error")
	prometheus.MustRegister(numberOfErrors)
}

func LogError(s string) {
	numberOfErrors.Inc()
	logrus.Error(s)
}
