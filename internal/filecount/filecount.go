package filecount

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rogierlommers/home/internal/config"
	"github.com/rogierlommers/home/internal/prom_error"
)

var pollInterval = 3600

func NewFileCounter(router *gin.Engine, cfg config.AppConfig) {

	// Start with defining metrics
	var filesInShareDrive = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "fileCount_ShareDrive",
		Help: "Number of files in the ShareDrive directory",
	})

	var filesInShareTMP = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "fileCount_ShareTMP",
		Help: "Number of files in the ShareDrive directory",
	})

	// Register the summary and the histogram with Prometheus's default registry
	prometheus.MustRegister(filesInShareDrive)
	prometheus.MustRegister(filesInShareTMP)

	go func() {
		for {

			// do the actual counting
			err := countFiles(cfg.FileCounterDrive, filesInShareDrive)
			if err != nil {
				prom_error.LogError(fmt.Sprintf("error counting fileCounterDrive: %s", err))
				time.Sleep(300 * time.Second)
			}

			err = countFiles(cfg.FileCounterTMP, filesInShareTMP)
			if err != nil {
				prom_error.LogError(fmt.Sprintf("error counting fileCounterTMP: %s", err))
				time.Sleep(300 * time.Second)
			}

			time.Sleep(time.Duration(pollInterval) * time.Second)
		}
	}()

}

func countFiles(d string, gauge prometheus.Gauge) error {
	var counter int

	err := filepath.Walk(d, func(path string, f os.FileInfo, err error) error {
		counter++
		return nil
	})

	if err != nil {
		return err
	}

	gauge.Set(float64(counter))
	return nil
}
