package handlers

import (
	"encoding/json"
	"flasher/config"
	"flasher/db"
	"net/http"
	//"path"
	"strconv"
	"time"
	"os" 
	"io"
	"path/filepath"

)

func HandleFlasher(cfg *config.Config, repo db.Repository) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		switch r.Method {

		case http.MethodGet:

			idStr := r.URL.Query().Get("id")
				osStr := r.URL.Query().Get("os")

				// Нельзя одновременно передавать id и os.
				if idStr != "" && osStr != "" {
					http.Error(
						w,
						"id and os cannot be used together",
						http.StatusBadRequest,
					)
					return
				}

				if idStr != "" {

					id, err := strconv.ParseInt(idStr, 10, 64)
					if err != nil {
						http.Error(
							w,
							"bad id",
							http.StatusBadRequest,
						)
						return
					}

					flasher, err := repo.GetFlasher(id)
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
						flasher.Path,
					)

					return
				}
 
				if osStr != "" {

					current, err := repo.GetCurrentFlasher(db.OSType(osStr))
					if err != nil {
						http.Error(
							w,
							err.Error(),
							http.StatusNotFound,
						)
						return
					}

					w.Header().Set(
						"Content-Type",
						"application/json",
					)

					json.NewEncoder(w).Encode(current)
					return
				}

				// Ни id, ни os не переданы.
				http.Error(
					w,
					"id or os is required",
					http.StatusBadRequest,
				)
  
		case http.MethodPost:

			const maxUploadSize = 100 << 20 // 100 MB

			if err := r.ParseMultipartForm(maxUploadSize); err != nil {
				http.Error(
					w,
					"invalid multipart form",
					http.StatusBadRequest,
				)
				return
			}

			osType := r.FormValue("os")

			if osType == "" {
				http.Error(
					w,
					"os is required",
					http.StatusBadRequest,
				)
				return
			}

			file, header, err := r.FormFile("flasher")
			if err != nil {
				http.Error(
					w,
					"flasher file is required",
					http.StatusBadRequest,
				)
				return
			}

			defer file.Close()

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

			now := time.Now()

			// DD-MM-YY-HH-MM-SS
			dirName := now.Format("02-01-06-15-04-05")

			saveDir := filepath.Join(
				cfg.FlasherDir,
				dirName,
			)

			if err := os.MkdirAll(saveDir, 0755); err != nil {
				http.Error(
					w,
					err.Error(),
					http.StatusInternalServerError,
				)
				return
			}

			filename := filepath.Base(header.Filename)

			path := filepath.Join(
				saveDir,
				filename,
			)

			if err := os.WriteFile(path, raw, 0644); err != nil {
				http.Error(
					w,
					err.Error(),
					http.StatusInternalServerError,
				)
				return
			}

			record := db.FlasherRecord{
				Name:      filename,
				OS:        osType,
				Size:      int64(len(raw)),
				SHA256:    sha256sum,
				Path:      path,
				IsCurrent: false,
				CreatedAt: now, 
			}

			if err := repo.AddFlasher(record); err != nil {
				_ = os.Remove(path)
				_ = os.Remove(saveDir)

				http.Error(
					w,
					err.Error(),
					http.StatusInternalServerError,
				)
				return
			}

			w.Header().Set(
				"Content-Type",
				"application/json",
			)

			w.WriteHeader(http.StatusCreated)

			json.NewEncoder(w).Encode(map[string]any{
				"ok":   true,
				"file": filename,
				"path": path,
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

			if err := repo.DeleteFlasher(id); err != nil {
				http.Error(
					w,
					err.Error(),
					http.StatusInternalServerError,
				)
				return
			}

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
				OS      string `json:"os"`
				Current bool   `json:"current"`
			}

			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				http.Error(
					w,
					"bad request",
					http.StatusBadRequest,
				)
				return
			}

			if request.OS == "" {
				http.Error(
					w,
					"missing os",
					http.StatusBadRequest,
				)
				return
			}

			osType := db.OSType(request.OS)

			if request.Current {

				err = repo.SetCurrentFlasher(
					id,
					osType,
				)

			} else {

				err = repo.UnsetCurrentFlasher(
					osType,
				)
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