package homeassistant

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/feeds"
	"github.com/rogierlommers/home/internal/config"
	"github.com/rogierlommers/home/internal/mailer"
	"github.com/rogierlommers/home/internal/sqlitedb"

	"github.com/sirupsen/logrus"
)

// curl -X POST "https://home.lommers.org/api/home-assistant" -H "Content-Type: application/json" -d '{"entity": "haha","message": "hihi"}'
// curl -X POST "http://localhost:3000/api/home-assistant" -H "Content-Type: application/json" -d '{"entity": "haha","message": "hihi"}'

var memcach *Cache

type Message struct {
	Entity  string `json:"entity"`
	Message string `json:"message"`
	Added   time.Time
}

// NewClient initializes NewHomeAssistant routes
func NewClient(router *gin.Engine, cfg config.AppConfig, m *mailer.Mailer, db *sqlitedb.DB) {

	// first create in-memory cache with maximum size of x entries
	memcach = newCache(100)

	router.POST("/api/home-assistant", incomingMessage(m, cfg, db))
	router.GET("/api/home-assistant/feed", displayRSS(db))

}

// dumpRequestBody reads and logs the request body, then restores it for further processing
func dumpRequestBody(c *gin.Context) error {
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		logrus.Errorf("Failed to read request body: %v", err)
		return err
	}

	// Log the raw body
	logrus.Debugf("Request body: %s", string(bodyBytes))

	// Restore the body so it can be read again
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	return nil
}

// documentation here: https://www.home-assistant.io/integrations/rest_command
func incomingMessage(m *mailer.Mailer, cfg config.AppConfig, db *sqlitedb.DB) gin.HandlerFunc {
	return func(c *gin.Context) {

		// first dump body
		if err := dumpRequestBody(c); err != nil {
			logrus.Errorf("failed to read JSON: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read request body"})
			return
		}

		// unmarshall body into message
		var msg Message
		if err := c.ShouldBindJSON(&msg); err != nil {
			logrus.Errorf("failed to bind JSON: %v", err)
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

			logrus.Debugf("cached message: entity=%s, message=%s", a.(*Message).Entity, a.(*Message).Message)

			incomingEntity := a.(*Message).Entity
			incomingMessage := a.(*Message).Message
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
