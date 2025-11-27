package homeassistant

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/feeds"
	"github.com/rogierlommers/home/internal/config"
	"github.com/rogierlommers/home/internal/mailer"
	"github.com/rogierlommers/home/internal/sqlitedb"

	"github.com/sirupsen/logrus"
)

var memcach *Cache

type Message struct {
	Entity  string `json:"entity"`
	Message string `json:"message"`
	Added   time.Time
}

// NewClient initializes NewHomeAssistant routes
func NewClient(router *gin.Engine, cfg config.AppConfig, m *mailer.Mailer, db *sqlitedb.DB) {

	// first create in-memory cache with maximum size of x entries
	memcach = newCache(10)

	router.POST("/api/home-assistant", incomingMessage(m, cfg, db))
	router.GET("/api/home-assistant/feed", displayRSS(db))

}

// documentation here: https://www.home-assistant.io/integrations/rest_command
func incomingMessage(m *mailer.Mailer, cfg config.AppConfig, db *sqlitedb.DB) gin.HandlerFunc {
	return func(c *gin.Context) {

		// unmarshall body into message
		var msg Message
		if err := c.ShouldBindJSON(&msg); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		// add message to in-memory cache
		logrus.Debugf("incoming message from Home Assistant: entity=%s, message=%s", msg.Entity, msg.Message)
		msg.Added = time.Now()
		memcach.Add(&msg)

		// respond okay
		c.JSON(http.StatusOK, gin.H{"msg": "all fine!"})
	}
}

// displayRSS produces an RSS feed from stored Home Assistant messages
func displayRSS(db *sqlitedb.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		now := time.Now()

		// create the feed
		feed := &feeds.Feed{
			Title:       "Home-assistant / notifications",
			Link:        &feeds.Link{},
			Description: "All notifications coming from Home Assistant",
			Created:     now,
		}

		// load all articles
		messages := memcach.GetElements()

		// add articles to the feed
		var newItem *feeds.Item

		for _, a := range messages {
			incomingMessage := a.(*Message).Message
			incomingEntity := a.(*Message).Entity
			incomingTimestamp := a.(*Message).Added

			newItem = &feeds.Item{
				Title: fmt.Sprintf("%s / %s", incomingEntity, incomingMessage),
				// Link:    &feeds.Link{Href: a.URL},
				Created: incomingTimestamp,
				// Id:      strconv.Itoa(a.ID),
			}

			feed.Add(newItem)
		}

		rss, err := feed.ToAtom()
		if err != nil {
			logrus.Errorf("error while generating RSS feed: %s", err)
			c.IndentedJSON(500, gin.H{"error": err.Error()})
			return
		}

		c.Writer.Write([]byte(rss))

	}
}
