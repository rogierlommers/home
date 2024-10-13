package message_webhook

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/rand"
)

type message struct {
	Timestamp time.Time
	Message   string
	ID        string
}

type incomingMessage struct {
	Data struct {
		Actor      string `json:"actor"`
		DeviceName string `json:"deviceName"`
		ManagedBy  string `json:"managedBy"`
		NodeID     string `json:"nodeID"`
		URL        string `json:"url"`
	} `json:"data"`
	Message   string    `json:"message"`
	Tailnet   string    `json:"tailnet"`
	Timestamp time.Time `json:"timestamp"`
	Type      string    `json:"type"`
	Version   int       `json:"version"`
}

func addMessage(c *gin.Context) {

	// read incoming message
	var f incomingMessage
	err := c.ShouldBindJSON(&f)
	if err != nil {
		logrus.Errorf("unable to read incoming data into struct: %s", err)
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// add message to in-memory queue
	x := message{
		Timestamp: time.Now(),
		Message:   f.Message,
		ID:        randomString(10),
	}
	cache.Add(x)

	// log and okay
	c.IndentedJSON(http.StatusCreated, x)
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randomString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}