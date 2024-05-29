package quicknote

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/sirupsen/logrus"
)

// response type to frontend
type response struct {
	Msg string `json:"msg"`
}

func sendMailHandler(c *gin.Context) {

	// read attachment; pure text will be added as .txt file
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		logrus.Errorf("error parsing formFile: %s", err)
		return
	}
	defer file.Close()

	// read file into buffer
	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, file); err != nil {
		logrus.Errorf("error reading file into buffer: %s", err)
		return
	}

	// mail the file
	if err := sendMail(header.Filename, buf.Bytes()); err != nil {
		c.JSON(http.StatusInternalServerError, response{Msg: fmt.Sprintf("error: mail error: %s", err.Error())})
		return
	}

	// write happy flow response
	c.JSON(200, response{Msg: "email succesfully sent"})
}
