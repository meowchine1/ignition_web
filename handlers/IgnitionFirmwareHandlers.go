package handlers

import (
	"flasher/db" 
	"net/http"
	"flasher/config" 
	"log"
	"os"
	"fmt"
	"io"
	"encoding/json"
	"strings"
	"path/filepath"
	"time" 
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/aes"
	"crypto/cipher"
	"encoding/binary" 
)
 

func computeHMAC(data []byte, cfg *config.Config) []byte {
	h := hmac.New(sha256.New, cfg.HMACKey)
	h.Write(data)
	return h.Sum(nil)
}

func encryptAndSign(fw []byte, cfg *config.Config) ([]byte, error) {
	iv := make([]byte, 16)
	if _, err := rand.Read(iv); err != nil {
		return nil, err
	}
	nonce := iv[:8]
	initialValue := binary.BigEndian.Uint64(iv[8:])
	block, err := aes.NewCipher(cfg.AESKey)
	if err != nil {
		return nil, err
	}
	counterBlock := make([]byte, aes.BlockSize)
	copy(counterBlock[:8], nonce)
	binary.BigEndian.PutUint64(counterBlock[8:], initialValue)
	stream := cipher.NewCTR(block, counterBlock)
	encrypted := make([]byte, len(fw))
	stream.XORKeyStream(encrypted, fw)
	mac := computeHMAC(append(iv, encrypted...), cfg)
	result := append(mac, iv...)
	result = append(result, encrypted...)
	return result, nil
}
 

func handleListFirmwares(cfg *config.Config, database *db.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Type", "application/json")

		firmwares, err := database.ListFirmwares()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"error": err.Error(),
			})
			return
		}

		if firmwares == nil {
			firmwares = []db.FirmwareRecord{}
		}

		json.NewEncoder(w).Encode(firmwares)
	}
}


func handleDownloadFirmware (cfg *config.Config, database *db.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {  
	id := strings.TrimPrefix(r.URL.Path, "/api/firmwares/")
	id = strings.Trim(id, "/")
	if id == "" || strings.Contains(id, "..") {
		http.Error(w, "invalid id", 400)
		return
	}
	path := filepath.Join( cfg.FirmwaresDir, filepath.Base(id+".enc"))
	http.ServeFile(w, r, path)
 }
}

func handleDeleteFirmware (cfg *config.Config, database *db.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {  
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", 405)
		return
	}
	token := r.Header.Get("X-Admin-Token")

  fmt.Println("X-Admin-Token:", token)
	//adminToken := os.Getenv("ADMIN_TOKEN")
  adminToken := cfg.AdminToken
	if adminToken == "" || token != adminToken {
		w.WriteHeader(401)
		json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/api/admin/delete/")
	id = strings.Trim(id, "/")
	if id == "" || strings.Contains(id, "..") {
		http.Error(w, "invalid id", 400)
		return
	}
	path := filepath.Join( cfg.FirmwaresDir, filepath.Base(id+".enc"))
	if err := os.Remove(path); err != nil {
		w.WriteHeader(404)
		json.NewEncoder(w).Encode(map[string]string{"error": "not found"})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
 }
}
 

func handleUploadFirmware (cfg *config.Config, database *db.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) { 
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", 405)
		return
	}
	token := r.Header.Get("X-Admin-Token")
	//adminToken := os.Getenv("ADMIN_TOKEN")
  adminToken := cfg.AdminToken
	if adminToken == "" || token != adminToken {
		w.WriteHeader(401)
		json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
		return
	}
	r.ParseMultipartForm(10 << 20)
	file, header, err := r.FormFile("firmware")
	if err != nil {
		w.WriteHeader(400)
		json.NewEncoder(w).Encode(map[string]string{"error": "no file"})
		return
	}
	defer file.Close()
	if !strings.HasSuffix(header.Filename, ".bin") {
		w.WriteHeader(400)
		json.NewEncoder(w).Encode(map[string]string{"error": "only .bin allowed"})
		return
	}
	raw, err := io.ReadAll(file)
	if err != nil || len(raw) == 0 {
		w.WriteHeader(400)
		json.NewEncoder(w).Encode(map[string]string{"error": "empty file"})
		return
	}
	encrypted, err := encryptAndSign(raw, cfg)
	if err != nil {
		w.WriteHeader(500)
		json.NewEncoder(w).Encode(map[string]string{"error": "encryption failed"})
		return
	}
	encName := filepath.Base(header.Filename) + ".enc"
	savePath := filepath.Join( cfg.FirmwaresDir, encName)
	if err := os.WriteFile(savePath, encrypted, 0644); err != nil {
		w.WriteHeader(500)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	log.Printf("[%s] Uploaded: %s (%d bytes raw → %d encrypted)",
		time.Now().Format("15:04:05"), header.Filename, len(raw), len(encrypted))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"ok":   true,
		"file": encName,
		"size": len(encrypted),
	})
 }
} 