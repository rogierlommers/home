package homepage

import (
	"os"
	"path/filepath"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/rogierlommers/home/internal/config"
	"github.com/rogierlommers/home/internal/stats"
	"github.com/sirupsen/logrus"
)

func scheduleCleanup(cfg config.AppConfig, statsDB *stats.DB) {
	c := cron.New()

	// schedule to run every day at 15:00
	_, err := c.AddFunc("0 15 * * *", func() {
		// cleanup files older than cfg.CleanUpInDys days
		cleanupOldFiles(cfg.UploadTarget, time.Duration(cfg.CleanUpInDys)*24*time.Hour, statsDB)
	})
	if err != nil {
		logrus.Errorf("failed to schedule cleanup: %v", err)
		return
	}

	logrus.Infof("scheduled %d-daily cleanup of old files in %s", cfg.CleanUpInDys, cfg.UploadTarget)
	c.Start()
}

func cleanupOldFiles(dir string, maxAge time.Duration, statsDB *stats.DB) {
	now := time.Now()

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip error, continue
		}

		if info.IsDir() {
			return nil
		}

		if now.Sub(info.ModTime()) > maxAge {
			if removeErr := os.Remove(path); removeErr != nil {
				logrus.Errorf("failed to remove file %s: %v", path, removeErr)
				statsDB.IncrementEntry("cleanup_errors")
			} else {
				logrus.Infof("removed old file: %s", path)
				statsDB.IncrementEntry("cleanup_files_removed")
			}
		} else {
			logrus.Debugf("file %s is not old enough to delete", path)
		}
		return nil
	})

	if err != nil {
		logrus.Errorf("error during cleanup: %v", err)
	}
}
