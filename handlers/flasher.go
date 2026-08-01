package handlers

import (
	"encoding/json"
	"flasher/config"
	"flasher/db"
	"net/http"
	//"path"
	//"strconv"
	"os" 
	"io"
	"path/filepath"

)

func HandleFlasher(cfg *config.Config, repo db.Repository) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		switch r.Method {

		case http.MethodGet:

			id, err := getIDFromQuery(r)

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

			w.Header().Set(
				"Content-Type",
				"application/json",
			)

			json.NewEncoder(w).Encode(flasher)

		case http.MethodPost:

			const maxUploadSize = 100 << 20 // 100 MB

			if err := r.ParseMultipartForm(maxUploadSize); err != nil {
				http.Error(w, "invalid multipart form", http.StatusBadRequest)
				return
			}

			file, header, err := r.FormFile("flasher")
			if err != nil {
				http.Error(w, "flasher file is required", http.StatusBadRequest)
				return
			}
			defer file.Close()

			if err := os.MkdirAll(cfg.FlasherDir, 0755); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			filename := filepath.Base(header.Filename)
			path := filepath.Join(cfg.FlasherDir, filename)

			dst, err := os.Create(path)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer dst.Close()

			if _, err := io.Copy(dst, file); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			record := db.FlasherRecord{
				Name: filename,
				Path: path,
			}

			if err := repo.AddFlasher(record); err != nil {
				os.Remove(path)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)

			json.NewEncoder(w).Encode(map[string]any{
				"ok":   true,
				"file": filename,
			})


			w.WriteHeader(http.StatusCreated)


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

		default:

			http.Error(
				w,
				"method not allowed",
				http.StatusMethodNotAllowed,
			)
		}
	}
}