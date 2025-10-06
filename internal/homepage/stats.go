package homepage

import (
	"github.com/gin-gonic/gin"
	"github.com/rogierlommers/home/internal/sqlitedb"
)

func statsHandler(statsDB *sqlitedb.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		counts, err := statsDB.GetAllEntryCounts()
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to get stats"})
			return
		}
		c.JSON(200, gin.H{"stats": counts})
	}
}
