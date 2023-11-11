package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/alexedwards/scs/postgresstore"
	"github.com/alexedwards/scs/v2"
	"github.com/bueti/shrinkster/internal/mailer"
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
	smtp struct {
		server   string
		port     int
		username string
		password string
		sender   string
	}
	signingKey string
	debug      bool
}

type application struct {
	config         config
	echo           *echo.Echo
	mailer         mailer.Mailer
	models         model.Models
	sessionManager *scs.SessionManager
}

func main() {
	var cfg config
	flag.IntVar(&cfg.port, "port", 8080, "API server port")
	flag.StringVar(&cfg.db.dsn, "db-dsn", "", "PostgreSQL DSN")
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "PostgreSQL max connection idle time")
	flag.StringVar(&cfg.signingKey, "signing-key", "", "JWT signing key")
	flag.StringVar(&cfg.smtp.server, "smtp-server", "smtp-relay.sendinblue.com", "SMTP server")
	flag.IntVar(&cfg.smtp.port, "smtp-port", 587, "SMTP port")
	flag.StringVar(&cfg.smtp.username, "smtp-username", "bbu+shrink@ik.me", "SMTP username")
	flag.StringVar(&cfg.smtp.password, "smtp-password", "", "SMTP password")
	flag.StringVar(&cfg.smtp.sender, "smtp-sender", "no-reply@shrink.ch", "SMTP sender")

	displayVersion := flag.Bool("version", false, "Display version and exit")

	flag.Parse()

	if *displayVersion {
		fmt.Printf("Version:\t%s\n", version)
		os.Exit(0)
	}

	parsEnvVars(&cfg)

	db, err := openDB(cfg)
	if err != nil {
		log.Fatal(err)
	}

	err = db.AutoMigrate(
		&model.User{},
		&model.Role{},
		&model.Url{},
		&model.Session{},
		&model.Token{},
	)
	if err != nil {
		log.Fatal(err)
	}

	dbd, _ := db.DB()

	sessionManager := scs.New()
	sessionManager.Store = postgresstore.New(dbd)
	sessionManager.Lifetime = 7 * 24 * time.Hour

	app := &application{
		sessionManager: sessionManager,
		mailer:         mailer.New(cfg.smtp.server, cfg.smtp.port, cfg.smtp.username, cfg.smtp.password, cfg.smtp.sender),
	}

	app.echo = app.initEcho()
	app.models = model.NewModels(db)
	app.config = cfg

	app.registerMiddleware()
	app.registerRoutes()
	app.serve()

}

func parsEnvVars(cfg *config) {
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

	if cfg.smtp.server == "" {
		envSMTPServer := os.Getenv("SMTP_SERVER")
		if envSMTPServer == "" {
			log.Fatal("SMTP_SERVER is required")
		}
		cfg.smtp.server = envSMTPServer
	}

	if cfg.smtp.username == "" {
		envSMTPUsername := os.Getenv("SMTP_USERNAME")
		if envSMTPUsername == "" {
			log.Fatal("SMTP_USERNAME is required")
		}
		cfg.smtp.username = envSMTPUsername
	}

	if cfg.smtp.password == "" {
		envSMTPPassword := os.Getenv("SMTP_PASSWORD")
		if envSMTPPassword == "" {
			log.Fatal("SMTP_PASSWORD is required")
		}
		cfg.smtp.password = envSMTPPassword
	}

	if cfg.smtp.port == 0 {
		envSMTPPort := os.Getenv("SMTP_PORT")
		if envSMTPPort == "" {
			log.Fatal("SMTP_PORT is required")
		}
		port, err := strconv.Atoi(envSMTPPort)
		if err != nil {
			log.Fatal(err)
		}
		cfg.smtp.port = port
	}

	if cfg.smtp.sender == "" {
		envSMTPSender := os.Getenv("SMTP_SENDER")
		if envSMTPSender == "" {
			log.Fatal("SMTP_SENDER is required")
		}
		cfg.smtp.sender = envSMTPSender
	}

	_, cfg.debug = os.LookupEnv("DEBUG")
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
