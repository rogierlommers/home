package homepage

import (
	"embed"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rogierlommers/home/internal/config"
	"github.com/rogierlommers/home/internal/mailer"
	"github.com/rogierlommers/home/internal/stats"
	"github.com/sirupsen/logrus"
)

var staticFS embed.FS

func Add(router *gin.Engine, cfg config.AppConfig, mailer *mailer.Mailer, staticHtmlFS embed.FS, statsDB *stats.DB) {

	// make the embedded filesystem available
	staticFS = staticHtmlFS

	// add routes
	router.GET("/", displayHome)
	router.POST("/api/upload", uploadFiles(cfg, mailer, statsDB))
	router.POST("/api/login", login(cfg))
	router.GET("/api/logout", logout())
	router.GET("/api/filelist", fileList(cfg))
	router.GET("/api/download", downloadFile(cfg))
	router.GET("/api/stats", statsHandler(statsDB))
	scheduleCleanup(cfg, statsDB)
}

func displayHome(c *gin.Context) {
	if !isAuthenticated(c) {
		htmlBytes, err := staticFS.ReadFile("static_html/login.html")
		if err != nil {
			c.String(500, "Failed to load login page")
			return
		}
		c.Header("Content-Type", "text/html")
		c.String(200, string(htmlBytes))
		return
	}

	htmlBytes, err := staticFS.ReadFile("static_html/homepage.html")
	if err != nil {
		c.String(500, "Failed to load homepage")
		return
	}
	c.Header("Content-Type", "text/html")
	c.String(200, string(htmlBytes))
}

func uploadFiles(cfg config.AppConfig, mailer *mailer.Mailer, stats *stats.DB) gin.HandlerFunc {
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

func isAuthenticated(c *gin.Context) bool {
	auth, err := c.Cookie("auth")
	return err == nil && auth == "true"
}

func login(cfg config.AppConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		var credentials struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}

		if err := c.BindJSON(&credentials); err != nil {
			logrus.Errorf("Failed to parse login request: %v", err)
			c.String(400, "Invalid request payload")
			return
		}

		if credentials.Username == cfg.Username && credentials.Password == cfg.Password {
			// valid for 6 months
			c.SetCookie("auth", "true", 15552000, "/", "", false, true)
			c.String(200, "Login successful")
			return
		} else {
			logrus.Errorf("Failed login attempt for user %s", credentials.Username)
			c.String(401, "Invalid username or password")
			return
		}
	}
}

func logout() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Clear the authentication cookie by setting its MaxAge to -1
		c.SetCookie("auth", "", -1, "/", "", false, true)
		c.Redirect(302, "/")
	}
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

func statsHandler(statsDB *stats.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		counts, err := statsDB.GetAllEntryCounts()
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to get stats"})
			return
		}
		c.JSON(200, gin.H{"stats": counts})
	}
}
