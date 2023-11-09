package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/bueti/shrinkster/internal/model"
	"github.com/labstack/echo/v4"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const version = "0.0.1"

type config struct {
	port int
	env  string
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
	}
	signingKey string
	debug      bool
}

type application struct {
	config config
	echo   *echo.Echo
	models model.Models
}

func main() {
	var cfg config
	flag.IntVar(&cfg.port, "port", 8080, "API server port")
	flag.StringVar(&cfg.db.dsn, "db-dsn", "", "PostgreSQL DSN")
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "PostgreSQL max connection idle time")
	flag.StringVar(&cfg.signingKey, "signing-key", "", "JWT signing key")

	displayVersion := flag.Bool("version", false, "Display version and exit")

	flag.Parse()

	if *displayVersion {
		fmt.Printf("Version:\t%s\n", version)
		os.Exit(0)
	}

	if cfg.db.dsn == "" {
		envDSN := os.Getenv("DSN")
		if envDSN == "" {
			log.Fatal("DSN is required")
		}
		cfg.db.dsn = envDSN
	}

	if cfg.signingKey == "" {
		envSigningKey := os.Getenv("SIGNING_KEY")
		if envSigningKey == "" {
			log.Fatal("SIGNING_KEY is required")
		}
		cfg.signingKey = envSigningKey
	}

	_, cfg.debug = os.LookupEnv("DEBUG")

	db, err := openDB(cfg)
	if err != nil {
		log.Fatal(err)
	}
	err = db.AutoMigrate(&model.User{})
	if err != nil {
		log.Fatal(err)
	}
	err = db.AutoMigrate(&model.Url{})
	if err != nil {
		log.Fatal(err)
	}
	err = db.AutoMigrate(&model.Role{})
	if err != nil {
		log.Fatal(err)
	}

	app := &application{}

	app.echo = app.initEcho()
	app.models = model.NewModels(db)
	app.config = cfg

	app.registerMiddleware()
	app.registerRoutes()
	app.serve()

}

func openDB(cfg config) (*gorm.DB, error) {
	loggerCfg := logger.Config{}
	if cfg.debug {
		loggerCfg.LogLevel = logger.Info
		loggerCfg.ParameterizedQueries = false
	} else {
		loggerCfg.LogLevel = logger.Error
		loggerCfg.ParameterizedQueries = true

	}
	loggerCfg.SlowThreshold = time.Second
	loggerCfg.Colorful = false

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		loggerCfg,
	)

	db, err := gorm.Open(postgres.Open(cfg.db.dsn), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(cfg.db.maxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.db.maxIdleConns)

	return db, nil
}
