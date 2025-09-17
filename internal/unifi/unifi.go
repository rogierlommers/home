package unifi

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rogierlommers/home/internal/config"
	"github.com/sirupsen/logrus"
)

type response struct {
	Msg  string `json:"msg"`
	Body string `json:"body"`
}

func NewUnifi(router *gin.Engine, cfg config.AppConfig) {
	router.POST("/api/unifi", unifiHandler)
	router.POST("/api/unifi/", unifiHandler)
}

func unifiHandler(c *gin.Context) {
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, response{Msg: "failed to read request body"})
		return
	}

	c.JSON(http.StatusCreated, response{
		Msg:  "email successfully sent",
		Body: string(bodyBytes),
	})

	logrus.Errorf("unifi webhook received: %s", string(bodyBytes))
}
