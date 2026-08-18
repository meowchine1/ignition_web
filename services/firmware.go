package services

import (
	"flasher/config"
	"flasher/db"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

type FirmwareService struct {
	cfg  *config.Config
	repo db.Repository
}

func NewFirmwareService(cfg *config.Config, repo db.Repository) *FirmwareService {
	return &FirmwareService{cfg: cfg, repo: repo}
}

// ---------- чтение ----------

func (s *FirmwareService) List() ([]db.FirmwareRecord, error) {
	firmwares, err := s.repo.ListFirmwares()
	if err != nil {
		return nil, err
	}
	if firmwares == nil {
		firmwares = []db.FirmwareRecord{}
	}
	return firmwares, nil
}

func (s *FirmwareService) ListAvailable() ([]db.FirmwareRecord, error) {
	all, err := s.List()
	if err != nil {
		return nil, err
	}

	available := make([]db.FirmwareRecord, 0, len(all))
	for _, fw := range all {
		if fw.IsAvailable {
			available = append(available, fw)
		}
	}
	return available, nil
}

func (s *FirmwareService) Get(id int) (db.FirmwareRecord, error) {
	fw, err := s.repo.GetFirmware(int64(id))
	if err != nil {
		return db.FirmwareRecord{}, err
	}
	return *fw, nil
}

func (s *FirmwareService) GetAvailable(id int) (db.FirmwareRecord, error) {
	fw, err := s.repo.GetFirmware(int64(id))
	if err != nil {
		return db.FirmwareRecord{}, err
	}
	if !fw.IsAvailable {
		return db.FirmwareRecord{}, fmt.Errorf("firmware is not available")
	}
	return *fw, nil
}

// ---------- запись ----------

func (s *FirmwareService) Create(originalName string, raw []byte) (db.FirmwareRecord, error) {
	if len(raw) == 0 {
		return db.FirmwareRecord{}, fmt.Errorf("empty file")
	}

	sha256sum := calculateSHA256(raw)

	encrypted, err := encryptAndSign(raw, s.cfg)
	if err != nil {
		return db.FirmwareRecord{}, fmt.Errorf("encryption failed: %w", err)
	}

	now := time.Now()

	// Создаем имя папки по времени: DD-MM-YY-HH-MM-SS
	dirName := now.Format("02-01-06-15-04-05")
	saveDir := filepath.Join(s.cfg.FirmwaresDir, dirName)

	if err := os.MkdirAll(saveDir, 0755); err != nil {
		return db.FirmwareRecord{}, fmt.Errorf("failed to create firmware directory: %w", err)
	}

	filename := originalName + ".enc"
	savePath := filepath.Join(saveDir, filename)

	if err := os.WriteFile(savePath, encrypted, 0644); err != nil {
		return db.FirmwareRecord{}, err
	}

	record := db.FirmwareRecord{
		Path:        savePath,
		CreatedAt:   now,
		SHA256:      sha256sum,
		IsAvailable: false,
		Size:        int64(len(encrypted)),
		Name:        originalName,
	}

	if err := s.repo.AddFirmware(record); err != nil {
		// Если запись в БД не добавилась — удаляем уже сохраненный файл.
		_ = os.Remove(savePath)
		_ = os.Remove(saveDir)
		return db.FirmwareRecord{}, err
	}

	log.Printf(
		"[%s] uploaded %s (%d -> %d bytes)",
		now.Format("15:04:05"),
		originalName,
		len(raw),
		len(encrypted),
	)

	return record, nil
}

func (s *FirmwareService) SetAvailable(id int, available bool) error {
	if available {
		return s.repo.EnableFirmware(int64(id))
	}
	return s.repo.DisableFirmware(int64(id))
}

func (s *FirmwareService) Delete(id int) error {
	firmware, err := s.repo.GetFirmware(int64(id))
	if err != nil {
		return err
	}

	if err := s.repo.DeleteFirmware(int64(id)); err != nil {
		return err
	}

	_ = os.Remove(firmware.Path)

	return nil
}

