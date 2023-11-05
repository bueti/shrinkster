package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/bueti/shrinkster/internal/model"
	"github.com/labstack/echo/v4"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
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
}

type application struct {
	config config
	echo   *echo.Echo
	db     *gorm.DB
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

	// initial echo router
	e := initEcho()

	app := &application{
		config: cfg,
		db:     db,
		echo:   e,
	}

	app.registerMiddleware()
	app.registerRoutes()
	app.serve()

}

func openDB(cfg config) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.db.dsn), &gorm.Config{})
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
