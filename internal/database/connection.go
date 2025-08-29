package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"dwell/internal/config"

	_ "github.com/lib/pq"
)

type Connection struct {
	DB *sql.DB
}

func NewConnection(cfg config.DatabaseConfig) (*Connection, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Successfully connected to database")

	return &Connection{DB: db}, nil
}

func (c *Connection) Close() error {
	return c.DB.Close()
}

func (c *Connection) GetDB() *sql.DB {
	return c.DB
}

