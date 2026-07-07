package db

import ( 
	"time"
)

type FileRecord struct {
	ID        int64
	Name      string
	Size      int64
	ModTime   time.Time
	SHA256    string
	Path      string
	CreatedAt time.Time
}

// =========================
// Firmware
// =========================

func (db *DB) AddFirmware(f FileRecord) error {

	_, err := db.Conn.Exec(
		`
		INSERT INTO firmware
		(
			name,
			size,
			mod_time,
			sha256,
			path
		)
		VALUES (?, ?, ?, ?, ?)
		`,
		f.Name,
		f.Size,
		f.ModTime,
		f.SHA256,
		f.Path,
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

func (db *DB) ListFirmwares() ([]FileRecord, error) {

	rows, err := db.Conn.Query(
		`
		SELECT
			id,
			name,
			size,
			mod_time,
			sha256,
			path,
			created_at
		FROM firmware
		ORDER BY created_at DESC
		`,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var result []FileRecord

	for rows.Next() {

		var f FileRecord

		err := rows.Scan(
			&f.ID,
			&f.Name,
			&f.Size,
			&f.ModTime,
			&f.SHA256,
			&f.Path,
			&f.CreatedAt,
		)

		if err != nil {
			return nil, err
		}

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

func (db *DB) AddFlasher(f FileRecord) error {

	_, err := db.Conn.Exec(
		`
		INSERT INTO flasher
		(
			name,
			size,
			mod_time,
			sha256,
			path
		)
		VALUES (?, ?, ?, ?, ?)
		`,
		f.Name,
		f.Size,
		f.ModTime,
		f.SHA256,
		f.Path,
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

func (db *DB) ListFlashers() ([]FileRecord, error) {

	rows, err := db.Conn.Query(
		`
		SELECT
			id,
			name,
			size,
			mod_time,
			sha256,
			path,
			created_at
		FROM flasher
		ORDER BY created_at DESC
		`,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var result []FileRecord

	for rows.Next() {

		var f FileRecord


		err := rows.Scan(
			&f.ID,
			&f.Name,
			&f.Size,
			&f.ModTime,
			&f.SHA256,
			&f.Path,
			&f.CreatedAt,
		)


		if err != nil {
			return nil, err
		}


		result = append(result, f)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}