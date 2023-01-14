package healthcheck

import (
	"net/http"

	"github.com/gin-gonic/gin"
	cfg "github.com/rogierlommers/quick-note/backend/config"
)

func AddRoutes(router *gin.Engine) {
	router.GET("/health", healthHandler)
}

func healthHandler(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, gin.H{"version": cfg.Settings.BackendVersion})
}
