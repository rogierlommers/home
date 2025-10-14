package homepage

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func displayNotify(c *gin.Context) {
	if !isAuthenticated(c) {
		c.Redirect(302, "/login")
		return
	}

	htmlBytes, err := staticFS.ReadFile("static_html/notify.html")
	if err != nil {
		logrus.Errorf("Error reading static html: %v", err)
		c.String(500, "Failed to load file notify page")
		return
	}

	c.Header("Content-Type", "text/html")
	c.String(200, string(htmlBytes))
}
