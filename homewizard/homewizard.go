package homewizard

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rogierlommers/home/config"
	"github.com/sirupsen/logrus"
)

var pollInterval = 60

type client struct {
	host   string
	client *http.Client
}

// Data data
type Data struct {
	SmrVersion            int64   `json:"smr_version"`
	MeterModel            string  `json:"meter_model"`
	WifiSSID              string  `json:"wifi_ssid"`
	WifiStrength          float64 `json:"wifi_strength"`
	TotalPowerImportT1Kwh float64 `json:"total_power_import_t1_kwh"`
	TotalPowerImportT2Kwh float64 `json:"total_power_import_t2_kwh"`
	TotalPowerExportT1Kwh float64 `json:"total_power_export_t1_kwh"`
	TotalPowerExportT2Kwh float64 `json:"total_power_export_t2_kwh"`
	ActivePowerW          float64 `json:"active_power_w"`
	ActivePowerL1W        float64 `json:"active_power_l1_w"`
	ActivePowerL2W        float64 `json:"active_power_l2_w"`
	ActivePowerL3W        float64 `json:"active_power_l3_w"`
	TotalGasM3            float64 `json:"total_gas_m3"`
}

func NewHomewizardExporter(router *gin.Engine, cfg config.AppConfig) {

	hwClient := &client{
		host: cfg.HomeWizardHost,
		client: &http.Client{
			Timeout: time.Second * 5,
		},
	}

	// Prometheus meters
	var (
		wifiStrength = promauto.NewGauge(prometheus.GaugeOpts{
			Name: "p1_wifi_strength",
			Help: "Wifi strength in Db",
		})

		totalPowerImportT1Kwh = promauto.NewGauge(prometheus.GaugeOpts{
			Name: "p1_total_power_import_t1_kwh",
			Help: "The total power import on T1 in kWh",
		})

		totalPowerImportT2Kwh = promauto.NewGauge(prometheus.GaugeOpts{
			Name: "p1_total_power_import_t2_kwh",
			Help: "The total power import on T2 in kWh",
		})

		totalPowerExportT1Kwh = promauto.NewGauge(prometheus.GaugeOpts{
			Name: "p1_total_power_export_t1_kwh",
			Help: "The total power export on T1 in kWh",
		})

		totalPowerExportT2Kwh = promauto.NewGauge(prometheus.GaugeOpts{
			Name: "p1_total_power_export_t2_kwh",
			Help: "The total power export on T2 in kWh",
		})

		activePowerW = promauto.NewGauge(prometheus.GaugeOpts{
			Name: "p1_active_power_w",
			Help: "The active power in W",
		})

		activePowerL1W = promauto.NewGauge(prometheus.GaugeOpts{
			Name: "p1_active_power_l1_w",
			Help: "he active power on L1 in W",
		})

		activePowerL2W = promauto.NewGauge(prometheus.GaugeOpts{
			Name: "p1_active_power_l2_w",
			Help: "he active power on L2 in W",
		})

		activePowerL3W = promauto.NewGauge(prometheus.GaugeOpts{
			Name: "p1_active_power_l3_w",
			Help: "The active power on L3 in W",
		})
	)

	go func() {
		for {
			home, err := hwClient.Retrieve()
			if err != nil {
				logrus.Errorf("error communicating with P1 meter: %s", err)
			}

			wifiStrength.Set(home.WifiStrength)

			totalPowerImportT1Kwh.Set(home.TotalPowerImportT1Kwh)
			totalPowerImportT2Kwh.Set(home.TotalPowerImportT2Kwh)
			totalPowerExportT1Kwh.Set(home.TotalPowerExportT2Kwh)
			totalPowerExportT2Kwh.Set(home.TotalPowerExportT2Kwh)

			activePowerW.Set(home.ActivePowerW)

			activePowerL1W.Set(home.ActivePowerL1W)
			activePowerL2W.Set(home.ActivePowerL2W)
			activePowerL3W.Set(home.ActivePowerL3W)

			time.Sleep(time.Duration(pollInterval) * time.Second)
		}
	}()

}

func (c *client) Retrieve() (home *Data, err error) {
	url := fmt.Sprintf("http://%s/api/v1/data", c.host)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		logrus.WithFields(logrus.Fields{"url": url}).Error("Coudln't create new http request", url, err)
		return nil, err
	}

	res, err := c.client.Do(req)
	if err != nil {
		logrus.Error("Couldn't execute request", err)
		return nil, err
	}
	if res.Body != nil {
		defer res.Body.Close()
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		logrus.WithFields(logrus.Fields{"body": body}).Error("Couldn't read body", err)
		return nil, err
	}

	home = &Data{}
	err = json.Unmarshal(body, &home)
	if err != nil {
		logrus.WithFields(logrus.Fields{"body": body}).Error("Couldn't parse body", err)
		return nil, err
	}

	return home, nil
}
