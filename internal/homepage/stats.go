package homepage

import (
	"github.com/gin-gonic/gin"
	"github.com/rogierlommers/home/internal/sqlitedb"
	"github.com/sirupsen/logrus"
)

func displayStatistics(c *gin.Context) {
	if !isAuthenticated(c) {
		c.Redirect(302, "/login")
		return
	}

	htmlBytes, err := staticFS.ReadFile("static_html/statistics.html")
	if err != nil {
		logrus.Errorf("Error reading static html: %v", err)
		c.String(500, "Failed to load homepage")
		return
	}

	c.Header("Content-Type", "text/html")
	c.String(200, string(htmlBytes))
}

func statsHandler(db *sqlitedb.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		counts, err := db.GetAllEntryCounts()
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to get stats"})
			return
		}
		c.JSON(200, gin.H{"stats": counts})
	}
}
