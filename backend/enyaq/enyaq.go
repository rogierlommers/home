package enyaq

import (
	"log"
	"time"

	"github.com/evcc-io/evcc/util"
	"github.com/evcc-io/evcc/vehicle/skoda"
	"github.com/evcc-io/evcc/vehicle/skoda/connect"
	"github.com/evcc-io/evcc/vehicle/vag/service"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var pollInterval = 60

func NewEnyaq(username string, password string, vin string) {

	var evRange = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "ev_range",
		Help: "Electric vehicle range",
	})

	var evSoc = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "ev_soc",
		Help: "Electric vehicle state of charge",
	})

	var evStatus = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "ev_status",
		Help: "Electric vehicle status",
	})

	var evFinishTime = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "ev_finish_time",
		Help: "Electric charging finish time",
	})

	var evOdometer = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "ev_odometer",
		Help: "Electric odometer",
	})

	// Register the summary and the histogram with Prometheus's default registry
	prometheus.MustRegister(evRange)
	prometheus.MustRegister(evSoc)
	prometheus.MustRegister(evStatus)
	prometheus.MustRegister(evFinishTime)
	prometheus.MustRegister(evOdometer)

	var err error

	logHandler := util.NewLogger("enyaq").Redact(username, password, vin)

	go func() {
		for {
			var provider *skoda.Provider
			if err == nil {
				ts, err := service.TokenRefreshServiceTokenSource(logHandler, skoda.TRSParams, connect.AuthParams, username, password)
				if err != nil {
					log.Print("TokenRefresh error: ", err)
					time.Sleep(time.Duration(pollInterval) * time.Second)
					continue
				}

				api := skoda.NewAPI(logHandler, ts)
				api.Client.Timeout = time.Second * 30

				provider = skoda.NewProvider(api, vin, time.Second*30)
			}

			rangeKm, err := provider.Range()
			if err != nil {
				log.Print("Range Error: ", err)
			} else {
				evRange.Set(float64(rangeKm))
			}

			soc, err := provider.Soc()
			if err != nil {
				log.Print("SoC error: ", err)
			} else {
				evSoc.Set(soc)
			}

			statusString, err := provider.Status()
			if err != nil {
				log.Print("Status error: ", err)
			} else {
				switch statusString.String() {
				case "":
					evStatus.Set(0)
				case "A":
					evStatus.Set(1)
				case "B":
					evStatus.Set(2)
				case "C":
					evStatus.Set(3)
				case "D":
					evStatus.Set(4)
				case "E":
					evStatus.Set(5)
				case "F":
					evStatus.Set(6)
				default:
					log.Print("Unknown status: ", statusString)
				}
			}

			finishTime, err := provider.FinishTime()
			if err != nil {
				log.Print("Finish time error: ", err)
			} else {
				evFinishTime.Set(float64(finishTime.Unix()))
			}

			odometer, err := provider.Odometer()
			if err != nil {
				log.Print("Odometer error: ", err)
			} else {
				evOdometer.Set(odometer)
			}

			time.Sleep(time.Duration(pollInterval) * time.Second)
		}
	}()

}

func AddRoutes(router *gin.Engine) {
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))
}
