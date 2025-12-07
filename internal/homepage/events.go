package homepage

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"text/template"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rogierlommers/home/internal/mailer"
	"github.com/rogierlommers/home/internal/sqlitedb"

	"github.com/sirupsen/logrus"
)

const eventsToRetrieve = 1000

// curl -X POST "https://home.lommers.org/api/events" -H "Content-Type: application/json" -d '{"source":"home-assistant", "label": "badkamer-boven","message": "hihi", "category": "sensor"}'
// curl -X POST "http://localhost:3000/api/events" -H "Content-Type: application/json" -d '{"source":"home-assistant", "label": "badkamer-boven","message": "hihi", "category": "sensor"}'

type Message struct {
	ID       int       `json:"id,omitempty"`
	Source   string    `json:"source"`
	Message  string    `json:"message"`
	Category string    `json:"category"`
	Added    time.Time `json:"added"`
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
func eventsIncomingMessage(m *mailer.Mailer, db *sqlitedb.DB) gin.HandlerFunc {
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

		// add to database
		if err := addEvent(db, msg); err != nil {
			logrus.Errorf("failed to add event to database: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store event"})
			return
		}

		// respond okay
		c.JSON(http.StatusOK, gin.H{"msg": "ok"})
	}
}

func serveEventsHTML() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !isAuthenticated(c) {
			c.Redirect(302, "/login")
			return
		}

		// Read the template file from the embedded FS
		htmlBytes, err := staticFS.ReadFile("static_html/events.html")
		if err != nil {
			logrus.Errorf("Error reading static html: %v", err)
			c.String(500, "Failed to load file events page")
			return
		}

		// Parse the template from bytes
		tmpl, err := template.New("events.html").Parse(string(htmlBytes))
		if err != nil {
			logrus.Errorf("Error parsing template: %v", err)
			c.String(500, "Failed to parse file events template")
			return
		}

		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, nil); err != nil {
			logrus.Errorf("Error executing template: %v", err)
			c.String(500, "Failed to render file storage page")
			return
		}

		c.Header("Content-Type", "text/html")
		c.String(200, buf.String())
	}
}

func addEvent(db *sqlitedb.DB, msg Message) error {
	_, err := db.Conn.Exec(`
		INSERT INTO events (source, message, category)
		VALUES (?, ?, ?)
	`, msg.Source, msg.Message, msg.Category)
	if err != nil {
		return fmt.Errorf("failed to insert event: %v", err)
	}
	return nil
}

func getEvents(db *sqlitedb.DB, number int, categoryFilter string) []Message {
	query := `SELECT id, message, category, added, source FROM events WHERE 1=1`
	args := []any{}

	if categoryFilter != "" {
		query += ` AND category = ?`
		args = append(args, categoryFilter)
	}

	query += ` ORDER BY id DESC LIMIT ?`
	args = append(args, number)

	rows, err := db.Conn.Query(query, args...)
	if err != nil {
		logrus.Errorf("failed to query events: %v", err)
		return nil
	}
	defer rows.Close()

	var events []Message
	for rows.Next() {
		var msg Message
		if err := rows.Scan(&msg.ID, &msg.Message, &msg.Category, &msg.Added, &msg.Source); err != nil {
			logrus.Errorf("failed to scan event row: %v", err)
			continue
		}
		events = append(events, msg)
	}
	if err := rows.Err(); err != nil {
		logrus.Errorf("row iteration error: %v", err)
	}
	return events
}

func displayEventsCategories(db *sqlitedb.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !isAuthenticated(c) {
			c.String(401, "Unauthorized")
			return
		}

		categories, err := db.GetEventsCategories()
		if err != nil {
			logrus.Errorf("Failed to get categories: %v", err)
			c.String(500, "Failed to retrieve categories")
			return
		}

		c.IndentedJSON(200, categories)
	}
}

func displayEvents(db *sqlitedb.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !isAuthenticated(c) {
			c.String(401, "Unauthorized")
			return
		}

		// get potential filter
		categoryFilter := c.Query("category")
		logrus.Debugf("category filter: %s", categoryFilter)

		events := getEvents(db, eventsToRetrieve, categoryFilter)
		c.IndentedJSON(200, events)
	}
}
