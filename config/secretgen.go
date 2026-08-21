package config

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"sort"
	"strings"
)

// InitSecrets генерирует отсутствующие machine secrets и дописывает их в .env.
// Генерация выполняется только один раз — если секрет уже существует,
// он НЕ перегенерируется. Возвращает (true, nil), если файл был изменён.
func InitSecrets() (bool, error) {
	envPath := ".env"

	// Проверяем, существует ли .env
	if _, err := os.Stat(envPath); err != nil && !os.IsNotExist(err) {
		return false, fmt.Errorf("failed to stat .env: %w", err)
	}

	// Machine secrets — генерируются один раз, не регенерируются при перезапуске.
	type secretDef struct {
		key  string
		desc string
		size int // размер в байтах
	}

	defs := []secretDef{
		{"JWT_SECRET", "JWT signing secret", 32},
		{"AES_KEY", "AES-128 encryption key", 16},
		{"HMAC_KEY", "HMAC-SHA256 key", 32},
	}

	// Не-секретные переменные со значениями по умолчанию, добавляются если отсутствуют
	fixed := []struct {
		key   string
		value string
	}{
		{"LISTEN_ADDR", "0.0.0.0:8200"},
		{"FIRMWARES_DIR", "./firmwares"},
		{"FLASHERS_DIR", "./flashers"},
	}

	changed := false

	// Читаем строки файла, чтобы сохранить порядок и комментарии
	originalLines := []string{}
	if data, err := os.ReadFile(envPath); err == nil {
		originalLines = strings.Split(strings.TrimRight(string(data), "\r\n"), "\n")
	}

	// Собираем новые строки
	var newLines []string
	seen := map[string]bool{}

	// Сначала пропускаем существующие строки, помечая ключи
	for _, line := range originalLines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			newLines = append(newLines, line)
			continue
		}
		parts := strings.SplitN(trimmed, "=", 2)
		if len(parts) != 2 {
			newLines = append(newLines, line)
			continue
		}
		key := strings.TrimSpace(parts[0])
		seen[key] = true
		newLines = append(newLines, line)
	}

	// Добавляем недостающие не-секретные переменные
	var toAddFixed []struct {
		key   string
		value string
	}
	for _, f := range fixed {
		if !seen[f.key] {
			toAddFixed = append(toAddFixed, f)
		}
	}
	sort.Slice(toAddFixed, func(i, j int) bool { return toAddFixed[i].key < toAddFixed[j].key })

	for _, f := range toAddFixed {
		newLines = append(newLines, fmt.Sprintf("%s=%s", f.key, f.value))
		seen[f.key] = true
		changed = true
	}

	// Добавляем недостающие секреты в алфавитном порядке для детерминизма
	var toAdd []secretDef
	for _, d := range defs {
		if !seen[d.key] {
			toAdd = append(toAdd, d)
		}
	}
	sort.Slice(toAdd, func(i, j int) bool { return toAdd[i].key < toAdd[j].key })

	for _, d := range toAdd {
		val, err := randomHexString(d.size)
		if err != nil {
			return false, fmt.Errorf("failed to generate %s: %w", d.key, err)
		}
		newLines = append(newLines, fmt.Sprintf("%s=%s  # %s", d.key, val, d.desc))
		changed = true
	}

	if !changed {
		return false, nil
	}

	// Приватный файл: только владелец может читать/писать
	if err := os.WriteFile(envPath, []byte(strings.Join(newLines, "\n")+"\n"), 0600); err != nil {
		return false, fmt.Errorf("failed to write .env: %w", err)
	}

	return true, nil
}

// randomHexString генерирует n байт случайных данных и кодирует их в hex.
func randomHexString(n int) (string, error) {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}