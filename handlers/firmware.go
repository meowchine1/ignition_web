package handlers

import (
	"flasher/config"
	"flasher/db"
	"net/http"
	//"path"
	//"strconv"
	"strings"
	"io" 
	"path/filepath"
	"os"
	"log"
	"time"
	"encoding/json" 

)

func HandleFirmware(cfg *config.Config, repo db.Repository) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		switch r.Method {

		case http.MethodGet:

			id, err := getIDFromQuery(r)

			if err != nil {
				log.Printf("err = %d id = %d", err, id)
				http.Error(
					w,
					"bad id",
					http.StatusBadRequest,
				)
				return
			}


			firmware, err := repo.GetFirmware(id)
			if err != nil {
				http.Error(
					w,
					err.Error(),
					http.StatusNotFound,
				)
				return
			}


			http.ServeFile(
				w,
				r,
				firmware.Path,
			)
 
		case http.MethodPost:

			if err := r.ParseMultipartForm(10 << 20); err != nil {
				http.Error(
					w,
					"bad form",
					http.StatusBadRequest,
				)
				return
			}

			file, header, err := r.FormFile("firmware")
			if err != nil {
				http.Error(
					w,
					"no file",
					http.StatusBadRequest,
				)
				return
			}

			defer file.Close()

			if !strings.HasSuffix(header.Filename, ".bin") {
				http.Error(
					w,
					"only .bin allowed",
					http.StatusBadRequest,
				)
				return
			}

			raw, err := io.ReadAll(file)
			if err != nil || len(raw) == 0 {
				http.Error(
					w,
					"empty file",
					http.StatusBadRequest,
				)
				return
			}

			sha256sum := calculateSHA256(raw)

			encrypted, err := encryptAndSign(raw, cfg)
			if err != nil {
				http.Error(
					w,
					"encryption failed",
					http.StatusInternalServerError,
				)
				return
			}

			now := time.Now()

			originalName := filepath.Base(header.Filename)
			filename := originalName + ".enc"

			// Создаем имя папки по времени:
			// DD-MM-YY-HH-MM-SS
			dirName := now.Format("02-01-06-15-04-05")

			saveDir := filepath.Join(
				cfg.FirmwaresDir,
				dirName,
			)

			if err := os.MkdirAll(saveDir, 0755); err != nil {
				http.Error(
					w,
					"failed to create firmware directory",
					http.StatusInternalServerError,
				)
				return
			}

			savePath := filepath.Join(
				saveDir,
				filename,
			)

			if err := os.WriteFile(savePath, encrypted, 0644); err != nil {
				http.Error(
					w,
					err.Error(),
					http.StatusInternalServerError,
				)
				return
			}

			record := db.FirmwareRecord{
				Path:      savePath,
				CreatedAt: now,
				SHA256:    sha256sum,
				IsAvailable : false,
				Size:      int64(len(encrypted)),
				Name:      originalName,
			}

			if err := repo.AddFirmware(record); err != nil {
				// Если запись в БД не добавилась,
				// удаляем уже сохраненный файл.
				_ = os.Remove(savePath)
				_ = os.Remove(saveDir)

				http.Error(
					w,
					err.Error(),
					http.StatusInternalServerError,
				)
				return
			}

			log.Printf(
				"[%s] uploaded %s (%d -> %d bytes)",
				now.Format("15:04:05"),
				header.Filename,
				len(raw),
				len(encrypted),
			)

			w.Header().Set(
				"Content-Type",
				"application/json",
			)

			w.WriteHeader(http.StatusCreated)

			json.NewEncoder(w).Encode(map[string]any{
				"ok":   true,
				"file": filename,
				"path": savePath,
			})


		case http.MethodDelete:

			id, err := getIDFromQuery(r)

			if err != nil {
				http.Error(
					w,
					"bad id",
					http.StatusBadRequest,
				)
				return
			}

			firmware, err := repo.GetFirmware(id)
			if err != nil {
				http.Error(
					w,
					"not found",
					http.StatusNotFound,
				)
				return
			}

			if err := repo.DeleteFirmware(id); err != nil {
				http.Error(
					w,
					err.Error(),
					http.StatusInternalServerError,
				)
				return
			}

			_ = os.Remove(firmware.Path)

			w.WriteHeader(http.StatusNoContent)
		
		case http.MethodPatch:

			id, err := getIDFromQuery(r)

			if err != nil {
				http.Error(
					w,
					"bad id",
					http.StatusBadRequest,
				)
				return
			}

			var request struct {
				Available bool `json:"available"`
			}

			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				http.Error(
					w,
					"bad request",
					http.StatusBadRequest,
				)
				return
			}

			if request.Available {
				err = repo.EnableFirmware(id)
			} else {
				err = repo.DisableFirmware(id)
			}

			if err != nil {
				http.Error(
					w,
					err.Error(),
					http.StatusInternalServerError,
				)
				return
			}

			w.WriteHeader(http.StatusNoContent)
			


		default:

			http.Error(
				w,
				"method not allowed",
				http.StatusMethodNotAllowed,
			)
		}
	}
}