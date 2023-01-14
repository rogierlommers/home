package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	cfg "github.com/rogierlommers/quick-note/backend/config"
	"github.com/rogierlommers/quick-note/backend/greedy"
	"github.com/rogierlommers/quick-note/backend/mailer"
)

// https://stackoverflow.com/questions/34046194/how-to-pass-arguments-to-router-handlers-in-golang-using-gin-web-framework

func AddRoutes(router *gin.Engine, m mailer.Mailer, g greedy.Greedy) {

	router.POST("/api/send", sendMailHandler(m))
	router.GET("/api/info", sendInfoHandler)
	router.Static("/static", cfg.Settings.StaticDir)

}

func sendMailHandler(m mailer.Mailer) gin.HandlerFunc {
	fn := func(c *gin.Context) {
		// incoming body
		type incoming struct {
			TodoItem string `json:"todo"`
		}

		// response type to frontend
		type response struct {
			Msg string `json:"msg"`
		}

		// marshal imcoming body into var
		var i incoming

		// Try to decode the request into the thumbnailRequest struct.
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

	return gin.HandlerFunc(fn)
}

func sendInfoHandler(c *gin.Context) {

	type response struct {
		Version string
	}

	c.JSON(200, response{
		Version: cfg.Settings.BackendVersion,
	})

}
