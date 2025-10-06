package homepage

import (
	"github.com/gin-gonic/gin"
	"github.com/rogierlommers/home/internal/sqlitedb"

	"github.com/sirupsen/logrus"
)

func displayBookmarks(db *sqlitedb.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !isAuthenticated(c) {
			c.String(401, "Unauthorized")
			return
		}

		bookmarks, err := db.GetBookmarks()
		if err != nil {
			logrus.Errorf("Failed to get bookmarks: %v", err)
			c.String(500, "Failed to retrieve bookmarks")
			return
		}

		c.IndentedJSON(200, bookmarks)
	}
}

func displayCategories(db *sqlitedb.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !isAuthenticated(c) {
			c.String(401, "Unauthorized")
			return
		}

		categories, err := db.GetCategories()
		if err != nil {
			logrus.Errorf("Failed to get categories: %v", err)
			c.String(500, "Failed to retrieve categories")
			return
		}

		c.IndentedJSON(200, categories)
	}
}

func addBookmark(db *sqlitedb.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !isAuthenticated(c) {
			c.String(401, "Unauthorized")
			return
		}
		var i sqlitedb.Item
		if err := c.BindJSON(&i); err != nil {
			logrus.Errorf("Failed to bind JSON: %v", err)
			c.String(400, "Invalid request payload")
			return
		}

		i.Type = "default" // hardcode for now
		logrus.Debugf("Received bookmark: %+v", i)
		if err := db.AddBookmark(i); err != nil {
			logrus.Errorf("Failed to add bookmark: %v", err)
			c.String(500, "Failed to add bookmark")
			return
		}

		logrus.Debug("Bookmark added successfully")
		c.IndentedJSON(201, i)
	}
}
