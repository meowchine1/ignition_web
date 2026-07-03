package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
  
  "flasher/config"
)

import _ "embed"

//go:embed ui/admin.html
var adminHTML string
 

type FirmwareInfo struct {
	Name    string `json:"name"`
	Size    int64  `json:"size"`
	ModTime string `json:"mod_time"`
}

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


func listFirmwares(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// cfg доступен здесь (захвачен из внешней функции)
		fmt.Println(cfg.FirmwaresDir)
//func listFirmwares(w http.ResponseWriter, r *http.Request, cfg *config.Config) {
	w.Header().Set("Content-Type", "application/json")
	entries, err := os.ReadDir(cfg.FirmwaresDir)
	if err != nil {
		w.WriteHeader(500)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	var list []FirmwareInfo
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".bin.enc") {
			continue
		}
		info, _ := e.Info()
		list = append(list, FirmwareInfo{
			Name:    strings.TrimSuffix(e.Name(), ".enc"),
			Size:    info.Size(),
			ModTime: info.ModTime().Format("2006-01-02 15:04"),
		})
	}
	if list == nil {
		list = []FirmwareInfo{}
	}
	json.NewEncoder(w).Encode(list)
 }
}

func downloadFirmware (cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) { // (w http.ResponseWriter, r *http.Request, cfg *config.Config) {
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

func deleteFirmware (cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) { // (w http.ResponseWriter, r *http.Request, cfg *config.Config) {
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

func uploadFirmware (cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) { //(w http.ResponseWriter, r *http.Request, cfg *config.Config) {
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

func adminPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, adminHTML)
}

func main() {

  fmt.Println("BUILD MARKER: 2026-07-03-NEW")
 
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	os.MkdirAll( cfg.FirmwaresDir, 0755)

	http.HandleFunc("/api/firmwares", listFirmwares(cfg))
	http.HandleFunc("/api/firmwares/", downloadFirmware(cfg))
	http.HandleFunc("/api/admin/upload", uploadFirmware(cfg))
	http.HandleFunc("/api/admin/delete/", deleteFirmware(cfg))
	http.HandleFunc("/", adminPage)

//fmt.Println(adminHTML)
  fmt.Println("len:", len(adminHTML))
fmt.Println("BUILD MARKER: 2026-07-03-NEW")
	fmt.Printf("IgnitionFlash Admin running on %s\n",  cfg.ListenAddr) 
	log.Fatal(http.ListenAndServe( cfg.ListenAddr, nil))
}

 