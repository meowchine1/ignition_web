package db

import (
	"database/sql"
	_ "embed"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed SQLite_schema.sql
var schemaSQL string


type SQLiteRepository struct {
	conn *sql.DB
}


// =========================
// Constructor
// =========================

func NewSQLiteRepository(path string) (Repository, error) {

	conn, err := sql.Open(
		"sqlite3",
		path,
	)

	if err != nil {
		return nil, err
	}

	cleanup := func() {
		_ = conn.Close()
	}
 
	conn.SetMaxOpenConns(1)
	conn.SetMaxIdleConns(1)
	conn.SetConnMaxLifetime(time.Hour)
 
	if err := conn.Ping(); err != nil {
		cleanup()
		return nil, err
	}
 
	if err := configureSQLite(conn); err != nil {
		cleanup()
		return nil, err
	}
 
	if err := initSchema(conn); err != nil {
		cleanup()
		return nil, err
	}
 
	return &SQLiteRepository{
		conn: conn,
	}, nil
}

 
func configureSQLite(conn *sql.DB) error {

	pragmas := []string{

		"PRAGMA foreign_keys = ON;",
		"PRAGMA journal_mode = WAL;",
		"PRAGMA synchronous = NORMAL;",
		"PRAGMA temp_store = MEMORY;",
		"PRAGMA mmap_size = 268435456;",
	}
 
	for _, pragma := range pragmas {

		if _, err := conn.Exec(pragma); err != nil {
			return err
		}
	} 
	return nil
}
 

func initSchema(conn *sql.DB) error {

	_, err := conn.Exec(schemaSQL)

	return err
}

func (r *SQLiteRepository) Close() error {
	return r.conn.Close()
}
 
// =========================
// Firmware
// =========================

func (r *SQLiteRepository) ClearCurrentFirmware() error {
	_, err := r.conn.Exec(
		`
		UPDATE firmware
		SET is_current = 0
		WHERE is_current = 1
		`,
	)

	return err
}

func (r *SQLiteRepository) UpdateFirmware(f FirmwareRecord) error {
	_, err := r.conn.Exec(
		`
		UPDATE firmware
		SET
			name = ? 
		WHERE id = ?
		`,
		f.Name, 
		f.ID,
	)

	return err
}

func (r *SQLiteRepository) SetCurrentFirmware(id int64) error {
	tx, err := r.conn.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	_, err = tx.Exec(
		`
		UPDATE firmware
		SET is_current = 0
		`,
	)

	if err != nil {
		return err
	}

	_, err = tx.Exec(
		`
		UPDATE firmware
		SET is_current = 1
		WHERE id = ?
		`,
		id,
	)

	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *SQLiteRepository) GetFirmware(id int64) (*FirmwareRecord, error) {
	var f FirmwareRecord

	err := r.conn.QueryRow(
		`
		SELECT
			id,
			name,
			size,
			version,
			sha256,
			path,
			created_at
		FROM firmware
		WHERE id = ?
		`,
		id,
	).Scan(
		&f.ID,
		&f.Name,
		&f.Size,
		&f.Version,
		&f.SHA256,
		&f.Path,
		&f.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &f, nil
}

func (r *SQLiteRepository) GetCurrentFirmware() (*FirmwareRecord, error) {
	var f FirmwareRecord
	var current int

	err := r.conn.QueryRow(
		`
		SELECT
			id,
			name,
			size,
			version,
			sha256,
			path,
			is_current,
			created_at
		FROM firmware
		WHERE is_current = 1
		LIMIT 1
		`,
	).Scan(
		&f.ID,
		&f.Name,
		&f.Size,
		&f.Version,
		&f.SHA256,
		&f.Path,
		&current,
		&f.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	f.IsCurrent = current == 1

	return &f, nil
}

func (r *SQLiteRepository) AddFirmware(f FirmwareRecord) error {

	_, err := r.conn.Exec(
		`
		INSERT INTO firmware
		(
			name,
			size,
			version,
			sha256,
			path,
			is_current
		)
		VALUES (?, ?, ?, ?, ?, ?)
		`,
		f.Name,
		f.Size,
		f.Version,
		f.SHA256,
		f.Path,
		boolToInt(f.IsCurrent),
	) 
	return err
}
 
func (r *SQLiteRepository) DeleteFirmware(id int64) error {

	_, err := r.conn.Exec(
		`
		DELETE FROM firmware
		WHERE id = ?
		`,
		id,
	) 
	return err
}
 
func (r *SQLiteRepository) ListFirmwares() ([]FirmwareRecord, error) {

	rows, err := r.conn.Query(
		`
		SELECT
			id,
			name,
			size,
			version,
			sha256,
			path,
			is_current,
			created_at
		FROM firmware
		ORDER BY created_at DESC
		`,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close() 
	var result []FirmwareRecord
 
	for rows.Next() {

		var f FirmwareRecord
		var current int
 
		err := rows.Scan(
			&f.ID,
			&f.Name,
			&f.Size,
			&f.Version,
			&f.SHA256,
			&f.Path,
			&current,
			&f.CreatedAt,
		)

		if err != nil {
			return nil, err
		}
 
		f.IsCurrent = current == 1

		result = append(result, f)
	}
 
	if err := rows.Err(); err != nil {
		return nil, err
	} 
	return result, nil
}
 
// =========================
// Flasher
// =========================

func (r *SQLiteRepository) SetCurrentFlasher(id int64, os OSType) error {

	tx, err := r.conn.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	_, err = tx.Exec(
		`
		UPDATE flasher
		SET is_current = 0
		WHERE os = ?
		`,
		os,
	)

	if err != nil {
		return err
	}
 
	_, err = tx.Exec(
		`
		UPDATE flasher
		SET is_current = 1
		WHERE id = ?
		`,
		id,
	)

	if err != nil {
		return err
	}
 
	return tx.Commit()
}

func (r *SQLiteRepository) ClearCurrentFlasher(os OSType) error {

	_, err := r.conn.Exec(
		`
		UPDATE flasher
		SET is_current = 0
		WHERE os = ?
		`,
		os,
	)

	return err
}

func (r *SQLiteRepository) UpdateFlasher(f FlasherRecord) error {
	_, err := r.conn.Exec(
		`
		UPDATE flasher
		SET
			name = ? 
		WHERE id = ?
		`,
		f.Name, 
		f.ID,
	)

	return err
}

func (r *SQLiteRepository) GetFlasher(id int64) (*FlasherRecord, error) {

	var f FlasherRecord
	var current int

	err := r.conn.QueryRow(
		`
		SELECT
			id,
			name,
			os,
			version,
			size,
			sha256,
			path,
			is_current,
			created_at
		FROM flasher
		WHERE id = ?
		`,
		id,
	).Scan(
		&f.ID,
		&f.Name,
		&f.OS,
		&f.Version,
		&f.Size,
		&f.SHA256,
		&f.Path,
		&current,
		&f.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	f.IsCurrent = current == 1

	return &f, nil
}

func (r *SQLiteRepository) GetCurrentFlasher(osType OSType) (*FlasherRecord, error) {
	var f FlasherRecord
	var current int

	err := r.conn.QueryRow(
		`
		SELECT
			id,
			name,
			os,
			version,
			size,
			sha256,
			path,
			is_current,
			created_at
		FROM flasher
		WHERE is_current = 1
		  AND os = ?
		LIMIT 1
		`,
		osType,
	).Scan(
		&f.ID,
		&f.Name,
		&f.OS,
		&f.Version,
		&f.Size,
		&f.SHA256,
		&f.Path,
		&current,
		&f.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	f.IsCurrent = current == 1

	return &f, nil
}

func (r *SQLiteRepository) AddFlasher(f FlasherRecord) error {

	_, err := r.conn.Exec(
		`
		INSERT INTO flasher
		(
			name,
			os,
			version,
			size,
			sha256,
			path,
			is_current
		)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		`,
		f.Name,
		f.OS,
		f.Version,
		f.Size,
		f.SHA256,
		f.Path,
		boolToInt(f.IsCurrent),
	) 
	return err
}
 
func (r *SQLiteRepository) DeleteFlasher(id int64) error {

	_, err := r.conn.Exec(
		`
		DELETE FROM flasher
		WHERE id = ?
		`,
		id,
	) 
	return err
}
 
func (r *SQLiteRepository) ListFlashers() ([]FlasherRecord, error) {

	rows, err := r.conn.Query(
		`
		SELECT
			id,
			name,
			os,
			version,
			size,
			sha256,
			path,
			is_current,
			created_at
		FROM flasher
		ORDER BY created_at DESC
		`,
	)
 
	if err != nil {
		return nil, err
	}
 
	defer rows.Close() 
	var result []FlasherRecord 
	for rows.Next() {

		var f FlasherRecord
		var current int
 
		err := rows.Scan(
			&f.ID,
			&f.Name,
			&f.OS,
			&f.Version,
			&f.Size,
			&f.SHA256,
			&f.Path,
			&current,
			&f.CreatedAt,
		)
 
		if err != nil {
			return nil, err
		}
 
		f.IsCurrent = current == 1
 
		result = append(result, f)
	}
 
	if err := rows.Err(); err != nil {
		return nil, err
	}
 
	return result, nil
}


// =========================
// Statistics
// =========================
func (r *SQLiteRepository) CountFirmwares() (int, error) {
	var count int

	err := r.conn.QueryRow(
		`
		SELECT COUNT(*)
		FROM firmware
		`,
	).Scan(&count)

	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *SQLiteRepository) CountFlashers() (int, error) {
	var count int

	err := r.conn.QueryRow(
		`
		SELECT COUNT(*)
		FROM flasher
		`,
	).Scan(&count)

	if err != nil {
		return 0, err
	}

	return count, nil
}
 
// =========================
// Helpers
// =========================

func boolToInt(v bool) int {

	if v {
		return 1
	}

	return 0
}