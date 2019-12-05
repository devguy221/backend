package main

import (
	"flag"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/zekroTJA/myrunes/internal/config"
	"github.com/zekroTJA/myrunes/internal/database"
	"github.com/zekroTJA/myrunes/internal/logger"
	"github.com/zekroTJA/myrunes/internal/webserver"
	"github.com/zekroTJA/myrunes/pkg/lifecycletimer"
)

var (
	flagConfig = flag.String("c", "config.yml", "config file location")
	flagAssets = flag.String("assets", "", "assets location")
)

func main() {
	flag.Parse()

	logger.Setup(`%{color}▶  %{level:.4s} %{id:03d}%{color:reset} %{message}`, 5)

	logger.Info("CONFIG :: initialization")
	cfg, err := config.Open(*flagConfig)
	if err != nil {
		logger.Fatal("CONFIG :: failed creating or opening config: %s", err.Error())
	}
	if cfg == nil {
		logger.Info("CONFIG :: config file was created at '%s'. Set your config values and restart.", *flagConfig)
		return
	}

	if v := os.Getenv("DB_HOST"); v != "" {
		cfg.MongoDB.Host = v
	}
	if v := os.Getenv("DB_PORT"); v != "" {
		cfg.MongoDB.Port = v
	}
	if v := os.Getenv("DB_USERNAME"); v != "" {
		cfg.MongoDB.Username = v
	}
	if v := os.Getenv("DB_PASSWORD"); v != "" {
		cfg.MongoDB.Password = v
	}
	if v := os.Getenv("DB_AUTHDB"); v != "" {
		cfg.MongoDB.AuthDB = v
	}
	if v := os.Getenv("DB_DATADB"); v != "" {
		cfg.MongoDB.DataDB = v
	}
	if v := strings.ToLower(os.Getenv("TLS_ENABLE")); v == "true" || v == "t" || v == "1" {
		cfg.WebServer.TLS.Enabled = true
	}
	if v := os.Getenv("TLS_KEY"); v != "" {
		cfg.WebServer.TLS.Key = v
	}
	if v := os.Getenv("TLS_CERT"); v != "" {
		cfg.WebServer.TLS.Cert = v
	}

	db := new(database.MongoDB)
	logger.Info("DATABASE :: initialization")
	if err = db.Connect(cfg.MongoDB); err != nil {
		logger.Fatal("DATABASE :: failed establishing connection to database: %s", err.Error())
	}
	defer func() {
		logger.Info("DATABASE :: teardown")
		db.Close()
	}()

	logger.Info("WEBSERVER :: initialization")
	ws := webserver.NewWebServer(db, cfg.WebServer, *flagAssets)
	go func() {
		if err := ws.ListenAndServeBlocking(); err != nil {
			logger.Fatal("WEBSERVER :: failed starting web server: %s", err.Error())
		}
	}()
	logger.Info("WEBSERVER :: started")

	lct := lifecycletimer.New(5 * time.Minute).
		Handle(func() {
			if err := db.CleanupExpiredSessions(); err != nil {
				logger.Error("DATABASE :: failed cleaning up sessions: %s", err.Error())
			}
		}).
		Start()
	defer lct.Stop()
	logger.Info("LIFECYCLETIMER :: started")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}
