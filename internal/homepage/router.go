package homepage

import (
	"embed"

	"github.com/gin-gonic/gin"
	"github.com/rogierlommers/home/internal/config"
	"github.com/rogierlommers/home/internal/mailer"
	"github.com/rogierlommers/home/internal/sqlitedb"
)

var staticFS embed.FS

func Add(router *gin.Engine, cfg config.AppConfig, mailer *mailer.Mailer, staticHtmlFS embed.FS, db *sqlitedb.DB) {

	// make the embedded filesystem available
	staticFS = staticHtmlFS

	// add routes
	router.GET("/", displayHome)
	router.POST("/api/upload", uploadFiles(cfg, mailer, db))
	router.POST("/api/login", login(cfg))
	router.GET("/api/logout", logout())
	router.GET("/api/filelist", fileList(cfg))
	router.GET("/api/download", downloadFile(cfg))
	router.GET("/api/stats", statsHandler(db))
	router.GET("/api/bookmarks", displayBookmarks(db, cfg.XHomeAPIKey))
	router.GET("/api/categories", displayCategories(db))
	router.POST("/api/bookmarks", addBookmark(db))

	scheduleCleanup(cfg, db)
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
