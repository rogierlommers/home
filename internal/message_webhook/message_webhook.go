package message_webhook

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rogierlommers/home/internal/config"
	"github.com/rogierlommers/tinycache"
	"github.com/sirupsen/logrus"
)

// curl -H 'Content-Type: application/json' -d '{ "title":"foo","body":"bar", "id": 1}' -X POST http://localhost:3000

var cache *tinycache.Cache

// define your custom struct, you can store everything you want
type message struct {
	timestamp time.Time
	message   string
}

func Add(router *gin.Engine, cfg config.AppConfig) {
	router.POST("/api/message_webhook", addMessage)
	cache = tinycache.NewCache(10)

	// use this for debugging purposes
	// go get github.com/tpkeeper/gin-dump

}

func addMessage(c *gin.Context) {
	logrus.Info("poep")
	x := message{
		timestamp: time.Now(),
		message:   "",
	}
	cache.Add(x)
}
