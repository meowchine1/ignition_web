package services

import (
	"flasher/config"
	"flasher/db"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type FlasherService struct {
	cfg  *config.Config
	repo db.Repository
}

func NewFlasherService(cfg *config.Config, repo db.Repository) *FlasherService {
	return &FlasherService{cfg: cfg, repo: repo}
}

// ---------- чтение ----------

func (s *FlasherService) List(osType string) ([]db.FlasherRecord, error) {
	var (
		flashers []db.FlasherRecord
		err      error
	)

	if osType == "" {
		flashers, err = s.repo.ListFlashers()
	} else {
		flashers, err = s.repo.ListFlashersByOS(db.OSType(osType))
	}

	if err != nil {
		return nil, err
	}
	if flashers == nil {
		flashers = []db.FlasherRecord{}
	}
	return flashers, nil
}

// Current — сохраняет старое поведение HandleFlasher(GET ?os=...):
// вернуть текущий (актуальный) флешер под указанную ОС.
func (s *FlasherService) Current(osType string) (db.FlasherRecord, error) {
	flasher, err := s.repo.GetCurrentFlasher(db.OSType(osType))
	if err != nil {
		return db.FlasherRecord{}, err
	}
	return *flasher, nil
}

func (s *FlasherService) Get(id int64) (db.FlasherRecord, error) {
	flasher, err := s.repo.GetFlasher(id)
	if err != nil {
		return db.FlasherRecord{}, err
	}
	return *flasher, nil
}

// ---------- запись ----------

func (s *FlasherService) Create(osType string, originalFilename string, raw []byte) (db.FlasherRecord, error) {
	if osType == "" {
		return db.FlasherRecord{}, fmt.Errorf("os is required")
	}
	if len(raw) == 0 {
		return db.FlasherRecord{}, fmt.Errorf("empty file")
	}

	sha256sum := calculateSHA256(raw)
	now := time.Now()

	// DD-MM-YY-HH-MM-SS
	dirName := now.Format("02-01-06-15-04-05")
	saveDir := filepath.Join(s.cfg.FlasherDir, dirName)

	if err := os.MkdirAll(saveDir, 0755); err != nil {
		return db.FlasherRecord{}, err
	}

	filename := filepath.Base(originalFilename)
	path := filepath.Join(saveDir, filename)

	if err := os.WriteFile(path, raw, 0644); err != nil {
		return db.FlasherRecord{}, err
	}

	record := db.FlasherRecord{
		Name:      filename,
		OS:        osType,
		Size:      int64(len(raw)),
		SHA256:    sha256sum,
		Path:      path,
		IsCurrent: false,
		CreatedAt: now,
	}

	if err := s.repo.AddFlasher(record); err != nil {
		_ = os.Remove(path)
		_ = os.Remove(saveDir)
		return db.FlasherRecord{}, err
	}

	return record, nil
}

func (s *FlasherService) SetCurrent(id int64, osType string, current bool) error {
	if osType == "" {
		return fmt.Errorf("missing os")
	}

	if current {
		return s.repo.SetCurrentFlasher(id, db.OSType(osType))
	}
	return s.repo.UnsetCurrentFlasher(db.OSType(osType))
}

func (s *FlasherService) Delete(id int64) error {
	return s.repo.DeleteFlasher(id)
}