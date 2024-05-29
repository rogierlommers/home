package enyaq

import (
	"fmt"
	"log"
	"time"

	"github.com/evcc-io/evcc/util"
	"github.com/evcc-io/evcc/vehicle/skoda"
	"github.com/evcc-io/evcc/vehicle/skoda/connect"
	"github.com/evcc-io/evcc/vehicle/vag/service"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rogierlommers/home/internal/config"
	"github.com/rogierlommers/home/internal/prom_error"
)

// | Value| Plugged-in | Charging | Explanation                                                                                  |
// |------|------------|----------|----------------------------------------------------------------------------------------------|
// |    0 |          ? |        ? | No status can be determined                                                                  |
// |    1 |          N |        N | Car is not plugged-in                                                                        |
// |    2 |          Y |        N | Car plugged-in but the vehicle is not charging                                               |
// |    3 |          Y |        Y | Car plugged-in and is charging                                                               |
// |    4 |          Y |        Y | Car plugged-in and charging, but with external ventilation request (for lead-acid batteries) |
// |    5 |          Y |        N | Car plugged-in, but not charging due to error: cable error (CP short circuit, 0V)            |
// |    6 |          Y |        N | Car plugged-in, but not charging. Simulate EVSE or unplugging error (CP wake-up, -12V)       |
// -------------------------------------------------------------------------------------------------------------------------------

var pollInterval = 60

func NewEnyaq(router *gin.Engine, cfg config.AppConfig) {

	// then start with metrics
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

	logHandler := util.NewLogger("enyaq").Redact(cfg.EnyaqUsername, cfg.EnyaqPassword, cfg.EnyaqVIN)

	go func() {
		for {
			var provider *skoda.Provider
			if err == nil {
				ts, err := service.TokenRefreshServiceTokenSource(logHandler, skoda.TRSParams, connect.AuthParams, cfg.EnyaqUsername, cfg.EnyaqPassword)
				if err != nil {
					prom_error.LogError(fmt.Sprintf("TokenRefresh error: %s", err))
					time.Sleep(time.Duration(pollInterval) * time.Second)
					continue
				}

				api := skoda.NewAPI(logHandler, ts)
				api.Client.Timeout = time.Second * 30

				provider = skoda.NewProvider(api, cfg.EnyaqVIN, time.Second*30)
			}

			rangeKm, err := provider.Range()
			if err != nil {
				prom_error.LogError(fmt.Sprintf("range error: %s", err))
			} else {
				evRange.Set(float64(rangeKm))
			}

			soc, err := provider.Soc()
			if err != nil {
				prom_error.LogError(fmt.Sprintf("soc error: %s", err))
			} else {
				evSoc.Set(soc)
			}

			statusString, err := provider.Status()
			if err != nil {
				prom_error.LogError(fmt.Sprintf("status error: %s", err))
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
			if err == nil {
				evFinishTime.Set(float64(finishTime.Unix()))
			}

			odometer, err := provider.Odometer()
			if err != nil {
				prom_error.LogError(fmt.Sprintf("odometer error: %s", err))
			} else {
				evOdometer.Set(odometer)
			}

			time.Sleep(time.Duration(pollInterval) * time.Second)
		}
	}()

}
