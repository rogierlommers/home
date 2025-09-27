package homepage

import (
	"os"
	"path/filepath"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/rogierlommers/home/internal/config"
	"github.com/rogierlommers/home/internal/mailer"
	"github.com/sirupsen/logrus"
)

func scheduleCleanup(cfg config.AppConfig, mailer *mailer.Mailer) {
	c := cron.New()

	// schedule to run every day at 15:00
	_, err := c.AddFunc("0 15 * * *", func() {
		cleanupOldFiles(cfg.UploadTarget, 30*24*time.Hour, mailer)
	})
	if err != nil {
		logrus.Errorf("failed to schedule cleanup: %v", err)
		return
	}

	logrus.Infof("scheduled %d-daily cleanup of old files in %s", cfg.CleanUpInDys, cfg.UploadTarget)
	c.Start()
}

func cleanupOldFiles(dir string, maxAge time.Duration, m *mailer.Mailer) {
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
			} else {
				logrus.Infof("removed old file: %s", path)
				subject := "Old file removed"
				body := "An old file has been removed from the server: " + info.Name()
				if mailErr := m.SendMail(subject, mailer.PrivateMail, body, nil); mailErr != nil {
					logrus.Errorf("failed to send cleanup notification email: %v", mailErr)
				}
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
