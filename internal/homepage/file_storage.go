package homepage

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
	"github.com/rogierlommers/home/internal/config"
	"github.com/rogierlommers/home/internal/mailer"
	"github.com/rogierlommers/home/internal/sqlitedb"
	"github.com/sirupsen/logrus"
)

func uploadFiles(cfg config.AppConfig, mailer *mailer.Mailer, stats *sqlitedb.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse the multipart form, with a max memory of 32 MB
		if err := c.Request.ParseMultipartForm(32 << 20); err != nil {
			c.String(400, "Failed to parse multipart form: %v", err)
			return
		}

		// Read the "notify" checkbox value from the form
		onlyUpload := c.PostForm("onlyUpload")
		logrus.Debugf("only upload: %s", onlyUpload)

		// Retrieve files from form data (multiple files)
		var uploaded []string
		var err error

		form := c.Request.MultipartForm
		files := form.File["file"]

		if len(files) == 0 {
			logrus.Debugf("No files uploaded")
		} else {
			uploaded, err = handleUploads(files, cfg)
			if err != nil {
				c.String(500, "Failed to upload files: %v", err)
				return
			}
		}

		// Create the uploads directory if it doesn't exist
		if _, err := os.Stat(cfg.UploadTarget); os.IsNotExist(err) {
			if err := os.MkdirAll(cfg.UploadTarget, os.ModePerm); err != nil {
				c.String(500, "Failed to create upload directory: %v", err)
				return
			}
		}

		// start sending here
		var subject, message, target, statsSource string

		if onlyUpload == "true" {
			statsSource = "upload_only_upload"
			logrus.Debugf("%d files uploaded without notification email", len(uploaded))

		} else {
			message = c.PostForm("message")
			subject = c.PostForm("subject")
			target = c.PostForm("targetEmail")

			logrus.Debugf("subject: %s", subject)
			logrus.Debugf("message: %s", message)
			logrus.Debugf("target: %s", target)

			// send mail
			statsSource = fmt.Sprintf("upload_and_notify_%s", target)
			if err := mailer.SendMail(subject, target, message, uploaded); err != nil {
				logrus.Errorf("Failed to send notification email: %v", err)
			} else {
				logrus.Info("Notification email sent")
			}
		}

		// increase stats
		if err := stats.IncrementEntry(statsSource); err != nil {
			logrus.Errorf("failed to increment upload_no_notify stat: %v", err)
		}

		c.String(200, "Files uploaded successfully: %v | notify: %s", uploaded, onlyUpload)
	}
}

func handleUploads(files []*multipart.FileHeader, cfg config.AppConfig) ([]string, error) {
	var uploaded []string
	for _, header := range files {
		file, err := header.Open()
		if err != nil {
			return nil, fmt.Errorf("failed to open file: %w", err)
		}
		defer file.Close()

		dstPath := filepath.Join(cfg.UploadTarget, header.Filename)
		out, err := os.Create(dstPath)
		if err != nil {
			return nil, fmt.Errorf("failed to create file: %v", err)
		}
		defer out.Close()

		if _, err := io.Copy(out, file); err != nil {
			return nil, fmt.Errorf("failed to save file: %v", err)
		}

		logrus.Debugf("uploaded file: %s", dstPath)
		uploaded = append(uploaded, header.Filename)
	}

	logrus.Debugf("received %d files for upload", len(files))
	return uploaded, nil
}

func fileList(cfg config.AppConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		entries, err := os.ReadDir(cfg.UploadTarget)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to list files"})
			return
		}

		type fileInfo struct {
			Name    string    `json:"name"`
			Size    int64     `json:"size"`
			ModTime time.Time `json:"modTime"`
		}
		var files []fileInfo
		for _, entry := range entries {
			if !entry.IsDir() {
				info, err := entry.Info()
				if err != nil {
					continue
				}
				files = append(files, fileInfo{
					Name:    entry.Name(),
					Size:    info.Size(),
					ModTime: info.ModTime(),
				})
			}
		}
		c.JSON(200, gin.H{"files": files})
	}
}

func downloadFile(cfg config.AppConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		logrus.Debugf("Download request from %s", c.ClientIP())
		filename := c.Query("filename")
		if filename == "" {
			c.String(400, "Filename query parameter is required")
			return
		}

		filePath := filepath.Join(cfg.UploadTarget, filename)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			c.String(404, "File not found")
			logrus.Errorf("file not found: %s", filePath)
			return
		}

		c.FileAttachment(filePath, filename)
	}
}

func scheduleCleanup(cfg config.AppConfig, statsDB *sqlitedb.DB) {
	c := cron.New()

	// schedule to run every day at 15:00
	_, err := c.AddFunc("0 15 * * *", func() {
		// cleanup files older than cfg.CleanUpInDys days
		cleanupOldFiles(cfg.UploadTarget, time.Duration(cfg.FileCleanUpInDys)*24*time.Hour, statsDB)
	})
	if err != nil {
		logrus.Errorf("failed to schedule cleanup: %v", err)
		return
	}

	logrus.Infof("scheduled %d-daily cleanup of old files in %s", cfg.FileCleanUpInDys, cfg.UploadTarget)
	c.Start()
}

func cleanupOldFiles(dir string, maxAge time.Duration, statsDB *sqlitedb.DB) {
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
