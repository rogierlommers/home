package main

import (
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rogierlommers/home/config"
	"github.com/rogierlommers/home/enyaq"
	"github.com/rogierlommers/home/greedy"
	"github.com/rogierlommers/home/homepage"
	"github.com/rogierlommers/home/homewizard"
	"github.com/rogierlommers/home/hue_exporter"
	"github.com/rogierlommers/home/quicknote"
	"github.com/sirupsen/logrus"
)

func main() {

	// disable logrus timestamp
	formatter := new(logrus.TextFormatter)
	formatter.DisableColors = true
	formatter.DisableTimestamp = true

	logrus.SetFormatter(formatter)

	// read config and make globally available
	config := config.ReadConfig()

	// create router
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	// add prometheus handler
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"POST", "PATCH"},
		AllowHeaders:     []string{"Access-Control-Allow-Headers, Origin,Accept, X-Requested-With, Content-Type, Access-Control-Request-Method, Access-Control-Request-Headers"},
		ExposeHeaders:    []string{"Content-Length"},
		MaxAge:           12 * time.Hour,
		AllowCredentials: true,
	}))

	// initialize all services
	homepage.Add(router, config)
	enyaq.NewEnyaq(router, config)
	quicknote.NewQuicknote(router, config)
	hue_exporter.NewHue(router, config)
	homewizard.NewHomewizardExporter(router, config)

	greedy, err := greedy.NewGreedy(config)
	if err != nil {
		logrus.Fatal(err)
	}
	defer greedy.CloseArticleDB()

	// schedule cleanup and routes
	greedy.AddRoutes(router)
	greedy.ScheduleCleanup()
	logrus.Infof("bucket initialized with %d records", greedy.Count())

	// show version number
	logrus.Info("version of: January 3 - 2024")

	// start serving
	if err := http.ListenAndServe(config.HostPort, router); err != nil {
		logrus.Fatal(err)
	}

}
