package main

import (
	"context"
	"database/sql"
	"flag"
	"html/template"
	"os"
	"time"

	_ "github.com/lib/pq"

	"github.com/0xrinful/LibraryMS/internal/data"
	"github.com/0xrinful/LibraryMS/internal/logger"
)

type config struct {
	port int
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
	}
}

type application struct {
	config        config
	logger        *logger.Logger
	models        data.Models
	templateCache map[string]*template.Template
}

func main() {
	cfg := parseFlags()
	logger := logger.New(os.Stdout, logger.LevelInfo)

	db, err := openDB(cfg)
	if err != nil {
		logger.PrintFatal(err)
	}
	defer db.Close()
	logger.PrintInfo("database connection pool established")

	cache, err := newTemplateCache()
	if err != nil {
		logger.PrintFatal(err)
	}

	app := &application{
		config:        cfg,
		logger:        logger,
		models:        data.NewModels(db),
		templateCache: cache,
	}

	err = app.serve()
	if err != nil {
		logger.PrintFatal(err)
	}
}

func openDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.db.maxOpenConns)
	db.SetMaxIdleConns(cfg.db.maxIdleConns)

	duration, err := time.ParseDuration(cfg.db.maxIdleTime)
	if err != nil {
		return nil, err
	}
	db.SetConnMaxIdleTime(duration)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}

	return db, nil
}

func parseFlags() config {
	var cfg config
	flag.IntVar(&cfg.port, "port", 8000, "Web Server port")
	flag.StringVar(&cfg.db.dsn, "db-dsn", "", "PostgreSQL DSN")
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.StringVar(
		&cfg.db.maxIdleTime,
		"db-max-idle-time",
		"15m",
		"PostgreSQL max connection idle time",
	)
	flag.Parse()
	return cfg
}
