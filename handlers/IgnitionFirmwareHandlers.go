package handlers

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary" 
	"flasher/config" 
	"net/http" 
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

func checkAdmin(r *http.Request, cfg *config.Config) bool {
	token := r.Header.Get("X-Admin-Token")
	return token != "" && token == cfg.AdminToken
}


// func HandleFirmwares(cfg *config.Config, database db.Repository) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {

// 		switch r.Method {

// 		case http.MethodGet:
// 			id := strings.TrimPrefix(r.URL.Path, "/api/firmwares")
// 			id = strings.Trim(id, "/")

// 			if id == "" {
// 				w.Header().Set("Content-Type", "application/json")

// 				firmwares, err := database.ListFirmwares()
// 				if err != nil {
// 					http.Error(w, err.Error(), 500)
// 					return
// 				}

// 				if firmwares == nil {
// 					firmwares = []db.FirmwareRecord{}
// 				}

// 				json.NewEncoder(w).Encode(firmwares)
// 				return
// 			}

// 			if strings.Contains(id, "..") {
// 				http.Error(w, "invalid id", 400)
// 				return
// 			}

// 			path := filepath.Join(
// 				cfg.FirmwaresDir,
// 				filepath.Base(id+".enc"),
// 			)

// 			http.ServeFile(w, r, path)


// 		case http.MethodPost:
// 			if !checkAdmin(r, cfg) {
// 				http.Error(w, "unauthorized", 401)
// 				return
// 			}

// 			if err := r.ParseMultipartForm(10 << 20); err != nil {
// 				http.Error(w, "bad form", 400)
// 				return
// 			}

// 			file, header, err := r.FormFile("firmware")
// 			if err != nil {
// 				http.Error(w, "no file", 400)
// 				return
// 			}

// 			defer file.Close()

// 			if !strings.HasSuffix(header.Filename, ".bin") {
// 				http.Error(w, "only .bin allowed", 400)
// 				return
// 			}

// 			raw, err := io.ReadAll(file)
// 			if err != nil || len(raw) == 0 {
// 				http.Error(w, "empty file", 400)
// 				return
// 			}

// 			encrypted, err := encryptAndSign(raw, cfg)
// 			if err != nil {
// 				http.Error(w, "encryption failed", 500)
// 				return
// 			}

// 			encName := filepath.Base(header.Filename) + ".enc"

// 			savePath := filepath.Join(
// 				cfg.FirmwaresDir,
// 				encName,
// 			)

// 			if err := os.WriteFile(savePath, encrypted, 0644); err != nil {
// 				http.Error(w, err.Error(), 500)
// 				return
// 			}

// 			log.Printf("[%s] Uploaded: %s (%d -> %d bytes)",
// 				time.Now().Format("15:04:05"),
// 				header.Filename,
// 				len(raw),
// 				len(encrypted),
// 			)

// 			json.NewEncoder(w).Encode(map[string]any{
// 				"ok":   true,
// 				"file": encName,
// 				"size": len(encrypted),
// 			})


// 		case http.MethodDelete:
// 			if !checkAdmin(r, cfg) {
// 				http.Error(w, "unauthorized", 401)
// 				return
// 			}

// 			id := strings.TrimPrefix(r.URL.Path, "/api/firmwares/")
// 			id = strings.Trim(id, "/")

// 			if id == "" || strings.Contains(id, "..") {
// 				http.Error(w, "invalid id", 400)
// 				return
// 			}

// 			path := filepath.Join(
// 				cfg.FirmwaresDir,
// 				filepath.Base(id+".enc"),
// 			)

// 			if err := os.Remove(path); err != nil {
// 				http.Error(w, "not found", 404)
// 				return
// 			}

// 			w.Header().Set("Content-Type", "application/json")
// 			json.NewEncoder(w).Encode(map[string]bool{
// 				"ok": true,
// 			})


// 		default:
// 			http.Error(w, "method not allowed", 405)
// 		}
// 	}
// }