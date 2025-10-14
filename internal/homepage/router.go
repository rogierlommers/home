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
	router.POST("/api/login", login(cfg))
	router.GET("/api/logout", logout())
	router.GET("/", func(c *gin.Context) {
		c.Redirect(302, "/bookmarks")
	})

	// statistics page
	router.GET("/api/stats", statsHandler(db))
	// router.GET("/statistics", displayStatistics)

	// bookmarks
	router.GET("/bookmarks", displayBookmarks)
	router.GET("/bookmarks/edit", displayEditBookmarks)
	router.GET("/api/bookmarks", getBookmarks(db, cfg.XHomeAPIKey))
	router.GET("/api/categories", displayCategories(db))
	router.POST("/api/bookmarks", addBookmark(db, cfg.XHomeAPIKey))
	router.PUT("/api/bookmarks/:id", editBookmark(db))
	router.DELETE("/api/bookmarks/:id", editBookmark(db))

	// file storage
	router.POST("/api/upload", uploadFiles(cfg, mailer, db))
	router.GET("/api/filelist", fileList(cfg))
	router.GET("/api/download", downloadFile(cfg))

	// cleanup
	scheduleCleanup(cfg, db)
}
