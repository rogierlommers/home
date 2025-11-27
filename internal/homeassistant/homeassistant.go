package homeassistant

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rogierlommers/home/internal/config"
	"github.com/rogierlommers/home/internal/mailer"
	"github.com/rogierlommers/home/internal/sqlitedb"
	"github.com/sirupsen/logrus"
)

// NewClient initializes NewHomeAssistant routes
func NewClient(router *gin.Engine, cfg config.AppConfig, m *mailer.Mailer, db *sqlitedb.DB) {

	router.POST("/api/home-assistant", homeAssistantEntities(m, cfg, db))

}

func homeAssistantEntities(m *mailer.Mailer, cfg config.AppConfig, db *sqlitedb.DB) gin.HandlerFunc {
	return func(c *gin.Context) {

		// log incoming request body
		body, err := c.GetRawData()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
			return
		}

		logrus.Debugf("Received Home Assistant webhook: %s", string(body))

		c.JSON(http.StatusOK, gin.H{"msg": "all fine!"})
	}
}
