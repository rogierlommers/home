package healthcheck

import (
	"net/http"

	"github.com/gin-gonic/gin"
	cfg "github.com/rogierlommers/quick-note/backend/config"
	"github.com/sirupsen/logrus"
)

func AddRoutes(router *gin.Engine) {
	router.GET("/health", healthHandler)
}

func healthHandler(c *gin.Context) {
	logrus.Info("incoming healthcheck")
	c.IndentedJSON(http.StatusOK, gin.H{"version": cfg.Settings.BackendVersion})
}
