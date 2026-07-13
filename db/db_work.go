package db

import (
	"time"
)


// =========================
// Firmware
// =========================

type FirmwareRecord struct {
	ID        int64
	Name      string
	Size      int64
	Version   string
	SHA256    string
	Path      string
	IsCurrent bool
	CreatedAt time.Time
}


func (db *DB) AddFirmware(f FirmwareRecord) error {

	_, err := db.Conn.Exec(
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


func (db *DB) DeleteFirmware(id int64) error {

	_, err := db.Conn.Exec(
		`
		DELETE FROM firmware
		WHERE id = ?
		`,
		id,
	)

	return err
}


func (db *DB) ListFirmwares() ([]FirmwareRecord, error) {

	rows, err := db.Conn.Query(
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

type FlasherRecord struct {
	ID        int64
	Name      string
	OS        string
	Version   string
	Size      int64
	SHA256    string
	Path      string
	IsCurrent bool
	CreatedAt time.Time
}



func (db *DB) AddFlasher(f FlasherRecord) error {

	_, err := db.Conn.Exec(
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



func (db *DB) DeleteFlasher(id int64) error {

	_, err := db.Conn.Exec(
		`
		DELETE FROM flasher
		WHERE id = ?
		`,
		id,
	)

	return err
}



func (db *DB) ListFlashers() ([]FlasherRecord, error) {

	rows, err := db.Conn.Query(
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
// Helpers
// =========================

func boolToInt(v bool) int {
	if v {
		return 1
	}

	return 0
}