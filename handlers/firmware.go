package handlers

import (
	"flasher/config"
	"flasher/db"
	"net/http"
	"path"
	"strconv"
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

			id, err := strconv.ParseInt(
				path.Base(r.URL.Path),
				10,
				64,
			)

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

			if !checkAdmin(r, cfg) {
				http.Error(
					w,
					"unauthorized",
					http.StatusUnauthorized,
				)
				return
			}


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


			encrypted, err := encryptAndSign(raw, cfg)
			if err != nil {
				http.Error(
					w,
					"encryption failed",
					http.StatusInternalServerError,
				)
				return
			}


			filename := filepath.Base(header.Filename) + ".enc"

			savePath := filepath.Join(
				cfg.FirmwaresDir,
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
				Path: savePath,
				// остальные поля по твоей структуре
			}


			if err := repo.AddFirmware(record); err != nil {
				http.Error(
					w,
					err.Error(),
					http.StatusInternalServerError,
				)
				return
			}


			log.Printf(
				"[%s] uploaded %s (%d -> %d bytes)",
				time.Now().Format("15:04:05"),
				header.Filename,
				len(raw),
				len(encrypted),
			)


			json.NewEncoder(w).Encode(map[string]any{
				"ok":   true,
				"file": filename,
			})


		case http.MethodDelete:

			id, err := strconv.ParseInt(
				path.Base(r.URL.Path),
				10,
				64,
			)

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


		default:

			http.Error(
				w,
				"method not allowed",
				http.StatusMethodNotAllowed,
			)
		}
	}
}