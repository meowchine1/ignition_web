package handlers

import (
	"encoding/json"
	"flasher/config"
	"flasher/db"
	"net/http"
)

func HandleCurrentFlasher(cfg *config.Config, repo db.Repository) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodGet {
			http.Error(
				w,
				"method not allowed",
				http.StatusMethodNotAllowed,
			)
			return
		}


		os := db.OSType(r.URL.Query().Get("os"))

		flasher, err := repo.GetCurrentFlasher(os)
		if err != nil {
			http.Error(
				w,
				err.Error(),
				http.StatusNotFound,
			)
			return
		}


		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(flasher)
	}
}