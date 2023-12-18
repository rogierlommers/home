package quicknote

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/sirupsen/logrus"
)

func sendMailHandler(c *gin.Context) {
	logrus.Info("incoming mail request")

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
		c.JSON(http.StatusInternalServerError, response{Msg: "error: EOF detected"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, response{Msg: err.Error()})
		return
	}

	// validate incoming message
	if len(i.TodoItem) == 0 {
		c.JSON(http.StatusInternalServerError, response{Msg: "error: empty todo-item detected"})
		return
	}

	// do something todo item
	if err := sendMail(i.TodoItem); err != nil {
		c.JSON(http.StatusInternalServerError, response{Msg: fmt.Sprintf("error: mail error: %s", err.Error())})
		return
	}

	// write happy flow response
	c.JSON(200, response{Msg: fmt.Sprintf("succesfully emailed: %d bytes", len(i.TodoItem))})
}
