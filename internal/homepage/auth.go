package homepage

import (
	"github.com/gin-gonic/gin"
	"github.com/rogierlommers/home/internal/config"
	"github.com/sirupsen/logrus"
)

func isAuthenticated(c *gin.Context) bool {
	auth, err := c.Cookie("auth")
	return err == nil && auth == "true"
}

func login(cfg config.AppConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		var credentials struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}

		if err := c.BindJSON(&credentials); err != nil {
			logrus.Errorf("Failed to parse login request: %v", err)
			c.String(400, "Invalid request payload")
			return
		}

		if credentials.Username == cfg.Username && credentials.Password == cfg.Password {
			// valid for 6 months
			c.SetCookie("auth", "true", 15552000, "/", "", false, true)
			c.String(200, "Login successful")
			return
		} else {
			logrus.Errorf("Failed login attempt for user %s", credentials.Username)
			c.String(401, "Invalid username or password")
			return
		}
	}
}

func logout() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Clear the authentication cookie by setting its MaxAge to -1
		c.SetCookie("auth", "", -1, "/", "", false, true)
		c.Redirect(302, "/")
	}
}
