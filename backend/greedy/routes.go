package greedy

import (
	"github.com/gin-gonic/gin"
)

func (g Greedy) AddRoutes(router *gin.Engine) {
	router.GET("/api/greedy/add", g.AddArticle)
	router.GET("/api/greedy/rss", g.DisplayRSS)
	router.GET("/api/greedy/accepted", g.AcceptedResponse)
}
