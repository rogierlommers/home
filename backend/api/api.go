package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	cfg "github.com/rogierlommers/quick-note/backend/config"
	"github.com/rogierlommers/quick-note/backend/mailer"
)

// type response struct {
// 	StatusText string `json:"status_text"`
// }

// https://stackoverflow.com/questions/34046194/how-to-pass-arguments-to-router-handlers-in-golang-using-gin-web-framework

func AddRoutes(router *gin.Engine, m mailer.Mailer) {

	router.POST("/api/send", sendMailHandler)
	router.GET("/api/info", sendInfoHandler)
	router.Static("/static", cfg.Settings.StaticDir)

}

func sendMailHandler(c *gin.Context) {

	// incoming body
	type incoming struct {
		TodoItem string `json:"todo"`
	}

	// response type ro requester
	type response struct {
		Msg string `json:"msg"`
	}

	// marshal imcoming body into var
	var i incoming

	// Try to decode the request into the thumbnailRequest struct.
	if err := json.NewDecoder(c.Request.Body).Decode(&i); err == io.EOF {
		c.JSON(http.StatusInternalServerError, response{Msg: "error: EOF detected"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	// validate incoming message
	if len(i.TodoItem) == 0 {
		c.JSON(http.StatusInternalServerError, response{Msg: "error: empty todo-item detected"})
		return
	}

	// do something todo item
	log.Printf("incoming todo item: %s", i.TodoItem)

	// write happy flow response
	c.JSON(200, response{Msg: fmt.Sprintf("succesfully emailed: %d bytes", len(i.TodoItem))})

}

func sendInfoHandler(c *gin.Context) {

	type response struct {
		Version string
	}

	c.JSON(200, response{
		Version: cfg.Settings.BackendVersion,
	})

}
