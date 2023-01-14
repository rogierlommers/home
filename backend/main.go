package main

import (
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	cfg "github.com/rogierlommers/quick-note/backend/config"
	"github.com/rogierlommers/quick-note/backend/greedy"
	"github.com/rogierlommers/quick-note/backend/mailer"
)

func main() {

	// disable logrus timestamp
	logrus.SetFormatter(new(logrus.TextFormatter))

	// read config and make globally available
	cfg.ReadConfig()

	// if mode is produciton, then tell it gin
	if cfg.Settings.Mode == "PRO" || cfg.Settings.Mode == "PRODUCTION" {
		logrus.Info("enabling production mode")
		gin.SetMode(gin.ReleaseMode)
	}

	// create mailer instance
	mailer := mailer.NewMailer()

	// create greedy instance
	greedy, err := greedy.NewGreedy()
	if err != nil {
		logrus.Fatal(err)
	}
	defer greedy.CloseArticleDB()

	// schedule cleanup
	greedy.ScheduleCleanup()
	logrus.Infof("bucket initialized with %d records", greedy.Count())

	// create router
	router := gin.New()
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"POST", "PATCH"},
		AllowHeaders:     []string{"Access-Control-Allow-Headers, Origin,Accept, X-Requested-With, Content-Type, Access-Control-Request-Method, Access-Control-Request-Headers"},
		ExposeHeaders:    []string{"Content-Length"},
		MaxAge:           12 * time.Hour,
		AllowCredentials: true,
	}))

	// add routes
	mailer.AddRoutes(router)
	greedy.AddRoutes(router)

	// start serving
	if err := http.ListenAndServe(":3000", router); err != nil {
		logrus.Fatal(err)
	}

}
