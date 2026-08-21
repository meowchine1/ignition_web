package config

import (
	"encoding/hex"
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

const DatabaseFileName = "database.sqlite3"

type Config struct {
	ListenAddr   string
	FirmwaresDir string
	FlasherDir   string
	DatabaseDir string
	AdminToken   string
	AESKey       []byte
	HMACKey      []byte

	JWTSecret string
}

// Load загружает конфигурацию из environment/.env.
// Секреты обязательны — если их нет, приложение не запускается.
// Обычный запуск никогда не генерирует секреты; для первоначальной
// генерации используйте команду: go run . -init-secrets
func Load() (*Config, error) {
	// .env может отсутствовать — переменные могут быть заданы в окружении.
	// godotenv не перезаписывает уже установленные переменные окружения.
	_ = godotenv.Load()

	cfg := &Config{
		ListenAddr:   getEnv("LISTEN_ADDR", "0.0.0.0:8200"),
		FirmwaresDir: getEnv("FIRMWARES_DIR", "./firmwares"),
		FlasherDir:   getEnv("FLASHERS_DIR", "./flashers"),
		DatabaseDir:  getEnv("DB_DIR", "./"),
		
		AdminToken:   getEnv("ADMIN_TOKEN", ""),
		JWTSecret:    getEnv("JWT_SECRET", ""),
	}

	// --- Обязательные секреты ---

	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required — set it in .env or run: go run . -init-secrets")
	}

	// --- AES-128 ключ (16 байт, hex) ---
	aesHex := getEnv("AES_KEY", "")
	if aesHex == "" {
		return nil, fmt.Errorf("AES_KEY is required — set it in .env or run: go run . -init-secrets")
	}
	aesKey, err := hex.DecodeString(aesHex)
	if err != nil {
		return nil, fmt.Errorf("AES_KEY must be hex-encoded: %w", err)
	}
	if len(aesKey) != 16 {
		return nil, fmt.Errorf("AES_KEY must be 16 bytes (32 hex chars), got %d bytes", len(aesKey))
	}
	cfg.AESKey = aesKey

	// --- HMAC-SHA256 ключ (32 байта, hex) ---
	hmacHex := getEnv("HMAC_KEY", "")
	if hmacHex == "" {
		return nil, fmt.Errorf("HMAC_KEY is required — set it in .env or run: go run . -init-secrets")
	}
	hmacKey, err := hex.DecodeString(hmacHex)
	if err != nil {
		return nil, fmt.Errorf("HMAC_KEY must be hex-encoded: %w", err)
	}
	if len(hmacKey) != 32 {
		return nil, fmt.Errorf("HMAC_KEY must be 32 bytes (64 hex chars), got %d bytes", len(hmacKey))
	}
	cfg.HMACKey = hmacKey

	return cfg, nil
}

// getEnv возвращает значение переменной окружения или fallback.
func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}