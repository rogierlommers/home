package mailer

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	cfg "github.com/rogierlommers/quick-note/backend/config"
)

func (m Mailer) AddRoutes(router *gin.Engine) {
	router.POST("/api/notes/send", m.sendMailHandler)
	router.Static("/static", cfg.Settings.StaticDir)
}

func (m Mailer) sendMailHandler(c *gin.Context) {
	// incoming body
	type incoming struct {
		TodoItem string `json:"todo"`
	}

	// response type to frontend
	type response struct {
		Msg string `json:"msg"`
	}

	// Try to decode the request into the thumbnailRequest struct.
	var i incoming
	if err := json.NewDecoder(c.Request.Body).Decode(&i); err == io.EOF {
		log.Println("test1")
		c.JSON(http.StatusInternalServerError, response{Msg: "error: EOF detected"})
		return
	} else if err != nil {
		log.Println("test2")
		c.JSON(http.StatusInternalServerError, response{Msg: err.Error()})
		return
	}

	// validate incoming message
	if len(i.TodoItem) == 0 {
		log.Println("test3")
		c.JSON(http.StatusInternalServerError, response{Msg: "error: empty todo-item detected"})
		return
	}

	// do something todo item
	if err := m.SendMail(i.TodoItem); err != nil {
		log.Println("test4")
		c.JSON(http.StatusInternalServerError, response{Msg: fmt.Sprintf("error: mail error: %s", err.Error())})
		return
	}

	// write happy flow response
	c.JSON(200, response{Msg: fmt.Sprintf("succesfully emailed: %d bytes", len(i.TodoItem))})
}
