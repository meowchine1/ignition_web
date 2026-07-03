package config

import (
	"encoding/hex"
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	ListenAddr   string
	FirmwaresDir string
	AdminToken   string
	AESKey       []byte
	HMACKey      []byte
}

func Load() (*Config, error) {

	// только для локальной разработки
	_ = godotenv.Load()

	cfg := &Config{
		ListenAddr:   os.Getenv("LISTEN_ADDR"),
		FirmwaresDir: os.Getenv("FIRMWARES_DIR"),
		AdminToken:   os.Getenv("ADMIN_TOKEN"),
	}

	if cfg.ListenAddr == "" {
		cfg.ListenAddr = "0.0.0.0:8080"
	}

	if cfg.FirmwaresDir == "" {
		cfg.FirmwaresDir = "./firmwares"
	}

	if cfg.AdminToken == "" {
		return nil, fmt.Errorf("ADMIN_TOKEN is required")
	}

	aesKey, err := hex.DecodeString(os.Getenv("AES_KEY"))
	if err != nil {
		return nil, fmt.Errorf("invalid AES_KEY: %w", err)
	}

	hmacKey, err := hex.DecodeString(os.Getenv("HMAC_KEY"))
	if err != nil {
		return nil, fmt.Errorf("invalid HMAC_KEY: %w", err)
	}

	cfg.AESKey = aesKey
	cfg.HMACKey = hmacKey

	return cfg, nil
}