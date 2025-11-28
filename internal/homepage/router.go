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

	// landing page and authorization
	router.GET("/", displayLoginPage)
	router.POST("/api/login", login(cfg))
	router.GET("/api/logout", logout())

	// statistics
	router.GET("/api/stats", statsHandler(db))
	router.GET("/statistics", displayStatistics)

	// bookmarks
	router.GET("/bookmarks", displayBookmarks)
	router.GET("/bookmarks/edit", displayEditBookmarks)
	router.GET("/api/bookmarks", getBookmarks(db, cfg.XHomeAPIKey))
	router.GET("/api/categories", displayCategories(db))
	router.GET("/api/bookmarks/export", createImportScript(db, cfg.XHomeAPIKey))
	router.POST("/api/bookmarks", addBookmark(db, cfg.XHomeAPIKey))
	router.PUT("/api/bookmarks/:id", editBookmark(db))
	router.DELETE("/api/bookmarks/:id", deleteBookmark(db))

	// notify
	router.GET("/notify", displayNotify)

	// file storage
	router.GET("/storage", displayStorage(cfg))
	router.GET("/api/filelist", fileList(cfg))
	router.GET("/api/download/:filename", downloadFile(cfg))
	router.POST("/api/upload", uploadFiles(cfg, mailer, db))

	// events
	router.GET("/events", serveEventsHTML(cfg))
	router.POST("/api/events", eventsIncomingMessage(mailer, db))
	router.GET("/api/events", displayEvents(db))
	router.GET("/api/events/categories", displayEventsCategories(db))
	router.GET("/api/events/labels", displayEventsLabels(db))

	// cleanup
	scheduleCleanup(cfg, db)
}
