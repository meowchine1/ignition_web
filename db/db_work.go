package db

import (
	"database/sql"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	Conn *sql.DB
}

func Init(path string) (*DB, error) {
	conn, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	// connection pool tuning
	conn.SetMaxOpenConns(1) // SQLite safe mode
	conn.SetMaxIdleConns(1)
	conn.SetConnMaxLifetime(time.Hour)

	if err := conn.Ping(); err != nil {
		return nil, err
	}

	// PRAGMA for performance + reliability
	pragmas := []string{
		"PRAGMA foreign_keys = ON;",
		"PRAGMA journal_mode = WAL;",
		"PRAGMA synchronous = NORMAL;",
		"PRAGMA temp_store = MEMORY;",
		"PRAGMA mmap_size = 268435456;", // 256MB
	}

	for _, p := range pragmas {
		if _, err := conn.Exec(p); err != nil {
			return nil, err
		}
	}

	if err := applyMigrations(conn); err != nil {
		return nil, err
	}

	return &DB{Conn: conn}, nil
}