package main

import (
	"embed"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/rogierlommers/home/internal/config"
	"github.com/rogierlommers/home/internal/greedy"
	"github.com/rogierlommers/home/internal/homepage"
	"github.com/rogierlommers/home/internal/mailer"
	"github.com/rogierlommers/home/internal/quicknote"
	"github.com/rogierlommers/home/internal/sqlitedb"
	"github.com/sirupsen/logrus"
)

//go:embed static_html/*
var staticHtmlFS embed.FS

func main() {

	// read config and make globally available
	cfg := config.ReadConfig()

	// read config and make globally available
	db := sqlitedb.InitDatabase(cfg)
	defer db.Close()

	// create router
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"POST", "PATCH"},
		AllowHeaders:     []string{"Access-Control-Allow-Headers, Origin,Accept, X-Requested-With, Content-Type, Access-Control-Request-Method, Access-Control-Request-Headers"},
		ExposeHeaders:    []string{"Content-Length"},
		MaxAge:           12 * time.Hour,
		AllowCredentials: true,
	}))

	// initialize all services
	mailer := mailer.NewMailer(cfg)
	homepage.Add(router, cfg, mailer, staticHtmlFS, db)
	quicknote.NewQuicknote(router, cfg, mailer, db)
	greedy.NewGreedy(router, cfg, db)

	// start serving
	logrus.Infof("listening on http://%s", cfg.HostPort)
	if err := http.ListenAndServe(cfg.HostPort, router); err != nil {
		logrus.Fatal(err)
	}

}
