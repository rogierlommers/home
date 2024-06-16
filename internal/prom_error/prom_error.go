package prom_error

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

// var numberOfErrors = prometheus.NewCounterVec(prometheus.CounterVecOpts{
// 	Name: "home_errors_total",
// 	Help: "Total number of errors occured in home app",
// 	[]string{"status", "environment", "reason"}})

var numberOfErrors = prometheus.NewCounterVec(prometheus.CounterOpts{
	Name: "home_errors_total",
	Help: "Total number of errors occured in home app",
}, []string{"package"})

func InitPromError() {
	logrus.Info("initialising prom error")
	prometheus.MustRegister(numberOfErrors)
}

func LogError(s string, key string) {
	numberOfErrors.WithLabelValues(key).Inc()
	logrus.Error(s)
}

func TriggerErrorHandler(router *gin.Engine) {
	router.GET("/api/error", justGenerateError)

}

func justGenerateError(c *gin.Context) {
	errString := "here you have an error!"
	LogError(errString, "test")
	c.JSON(http.StatusInternalServerError, errString)
}
