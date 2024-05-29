package main

import (
	"net/http"
	"time"

	"github.com/rogierlommers/home/internal/filecount"
	"github.com/rogierlommers/home/internal/prom_error"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rogierlommers/home/internal/config"
	"github.com/rogierlommers/home/internal/enyaq"
	"github.com/rogierlommers/home/internal/greedy"
	"github.com/rogierlommers/home/internal/homepage"
	"github.com/rogierlommers/home/internal/hue_exporter"
	"github.com/rogierlommers/home/internal/quicknote"
	"github.com/sirupsen/logrus"
)

func main() {

	// disable logrus timestamp
	formatter := new(logrus.TextFormatter)
	formatter.DisableColors = true
	formatter.DisableTimestamp = true

	logrus.SetFormatter(formatter)

	// init package for logging errors
	prom_error.InitPromError()

	// read config and make globally available
	cfg := config.ReadConfig()

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
	filecount.NewFileCounter(router, cfg)
	homepage.Add(router, cfg)
	enyaq.NewEnyaq(router, cfg)
	quicknote.NewQuicknote(router, cfg)
	hue_exporter.NewHue(router, cfg)

	greedyInstance, err := greedy.NewGreedy(cfg)
	if err != nil {
		logrus.Fatal(err)
	}
	defer greedyInstance.CloseArticleDB()

	// schedule cleanup and routes
	greedyInstance.AddRoutes(router)
	greedyInstance.ScheduleCleanup()
	logrus.Infof("bucket initialized with %d records", greedyInstance.Count())

	// show version number
	logrus.Info("version of: May 29 - 2024")

	// start serving
	logrus.Infof("listening on %s", cfg.HostPort)
	if err := http.ListenAndServe(cfg.HostPort, router); err != nil {
		logrus.Fatal(err)
	}

}
