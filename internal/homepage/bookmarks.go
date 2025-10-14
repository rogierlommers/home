package homepage

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rogierlommers/home/internal/sqlitedb"

	"github.com/sirupsen/logrus"
)

func displayBookmarks(c *gin.Context) {
	if !isAuthenticated(c) {
		c.Redirect(302, "/login")
		return
	}

	htmlBytes, err := staticFS.ReadFile("static_html/bookmarks.html")
	if err != nil {
		logrus.Errorf("Error reading static html: %v", err)
		c.String(500, "Failed to load homepage")
		return
	}

	c.Header("Content-Type", "text/html")
	c.String(200, string(htmlBytes))
}

func getBookmarks(db *sqlitedb.DB, XAPIkey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-HOME-API-KEY")
		if !isAuthenticated(c) && apiKey != XAPIkey {
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

		excludeHiddenStr := c.Query("exclude_hidden")
		excludeHidden, _ := strconv.ParseBool(excludeHiddenStr)

		categories, err := db.GetCategories(excludeHidden)
		if err != nil {
			logrus.Errorf("Failed to get categories: %v", err)
			c.String(500, "Failed to retrieve categories")
			return
		}

		c.IndentedJSON(200, categories)
	}
}

func addBookmark(db *sqlitedb.DB, XAPIkey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-HOME-API-KEY")
		if !isAuthenticated(c) && apiKey != XAPIkey {
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

func displayEditBookmarks(c *gin.Context) {
	if !isAuthenticated(c) {
		c.Redirect(302, "/login")
		return
	}

	htmlBytes, err := staticFS.ReadFile("static_html/bookmarks_edit.html")
	if err != nil {
		logrus.Errorf("Error reading static html: %v", err)
		c.String(500, "Failed to load homepage")
		return
	}

	c.Header("Content-Type", "text/html")
	c.String(200, string(htmlBytes))
}

func deleteBookmark(db *sqlitedb.DB) gin.HandlerFunc {
	return func(c *gin.Context) {

		id := convertToInt(c.Param("id"))

		if !isAuthenticated(c) {
			c.String(401, "Unauthorized")
			return
		}

		if err := db.DeleteBookmark(id); err != nil {
			logrus.Errorf("Failed to delete bookmark: %v", err)
			c.String(500, "Failed to delete bookmark")
			return
		}

		logrus.Debugf("Deleted bookmark ID %d", id)
		c.JSON(201, gin.H{"status": "ok", "deletedID": id})
	}
}

func editBookmark(db *sqlitedb.DB) gin.HandlerFunc {
	return func(c *gin.Context) {

		id := convertToInt(c.Param("id"))

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

		i.ID = id
		if err := db.UpdateBookmark(i); err != nil {
			logrus.Errorf("Failed to update bookmark: %v", err)
			c.String(500, "Failed to update bookmark")
			return
		}

		logrus.Debugf("Updated bookmark ID %d", id)
		c.JSON(201, i)
	}
}

func convertToInt(s string) int {
	var i int
	_, err := fmt.Sscanf(s, "%d", &i)
	if err != nil {
		logrus.Errorf("Error converting string to int: %v", err)
		return 0
	}
	return i
}
