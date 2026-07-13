package db

import (
	"database/sql" 
	"time" 
	_ "github.com/mattn/go-sqlite3"
)

import _ "embed"
//go:embed schema.sql
var schemaSQL string

type DB struct {
	Conn *sql.DB
}

func initSchema(conn *sql.DB) error {

	_, err := conn.Exec(schemaSQL)

	return err
	 
}

func Init() (*DB, error) {
	conn, err := sql.Open(
		"sqlite3",
		"./database.sqlite3",
	)

	if err != nil {
		return nil, err
	}

	conn.SetMaxOpenConns(1)
	conn.SetMaxIdleConns(1)
	conn.SetConnMaxLifetime(time.Hour)

	if err := conn.Ping(); err != nil {
		conn.Close()
		return nil, err
	}

	pragmas := []string{

		"PRAGMA foreign_keys = ON;",
		"PRAGMA journal_mode = WAL;",
		"PRAGMA synchronous = NORMAL;",
		"PRAGMA temp_store = MEMORY;",
		"PRAGMA mmap_size = 268435456;",
	}

	for _, p := range pragmas {

		if _, err := conn.Exec(p); err != nil {

			conn.Close()

			return nil, err
		}
	}

	if err := initSchema(conn); err != nil {

		conn.Close()

		return nil, err
	}

	return &DB{
		Conn: conn,
	}, nil
}


 