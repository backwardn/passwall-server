package main

import (
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/heroku/x/hmetrics/onload"
	"github.com/pass-wall/passwall-server/internal/app"
	"github.com/pass-wall/passwall-server/internal/config"
	"github.com/pass-wall/passwall-server/internal/router"
	"github.com/pass-wall/passwall-server/internal/storage"
)

func main() {
	logger := log.New(os.Stdout, "[passwall-server] ", 0)

	cfg, err := config.SetupConfigDefaults()
	if err != nil {
		log.Fatal(err)
	}

	db, err := storage.DBConn(&cfg.Database)
	if err != nil {
		log.Fatal(err)
	}

	s := storage.New(db)

	// Migrate database tables
	// TODO: Migrate should be in storege.New functions of categories
	app.MigrateSystemTables(s)

	// Start cron jobs like backup
	// app.StartCronJob(s)

	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		WriteTimeout: time.Second * time.Duration(cfg.Server.Timeout),
		ReadTimeout:  time.Second * time.Duration(cfg.Server.Timeout),
		IdleTimeout:  time.Second * 60,
		Handler:      router.New(s),
	}

	logger.Printf("listening on %s", cfg.Server.Port)
	if err := srv.ListenAndServe(); err != nil {
		logger.Fatal(err)
	}
}
