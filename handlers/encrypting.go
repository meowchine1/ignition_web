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
	"encoding/hex"
	"strconv"
)

func computeHMAC(data []byte, cfg *config.Config) []byte {
	h := hmac.New(sha256.New, cfg.HMACKey)
	h.Write(data)
	return h.Sum(nil)
}

func getIDFromQuery(r *http.Request) (int64, error) {
	idStr := r.URL.Query().Get("id")

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return 0, err
	}

	return id, nil
}


func calculateSHA256(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
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

 