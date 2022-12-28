package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type MailResponse struct {
	StatusText string `json:"status_text"`
}

type MailRequest struct {
	Text string `json:"text"`
}

func sendMailHandler(c *gin.Context) {
	var decodedRequest MailRequest

	// Try to decode the request into the thumbnailRequest struct.
	err := json.NewDecoder(c.Request.Body).Decode(&decodedRequest)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	// do something with decodedRequest

	// write response
	mailResponse := MailResponse{
		StatusText: fmt.Sprintf("succesfully emailed: %d bytes", len(decodedRequest.Text)),
	}

	c.JSON(200, mailResponse)
}

func main() {
	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"POST", "PATCH"},
		AllowHeaders:     []string{"Access-Control-Allow-Headers, Origin,Accept, X-Requested-With, Content-Type, Access-Control-Request-Method, Access-Control-Request-Headers"},
		ExposeHeaders:    []string{"Content-Length"},
		MaxAge:           12 * time.Hour,
		AllowCredentials: true,
	}))

	router.POST("/api/send", sendMailHandler)

	if err := http.ListenAndServe(":3000", router); err != nil {
		logrus.Fatal(err)
	}

}
