package handlers

import (
	"encoding/json"
	"flasher/config"
	"flasher/db"
	"net/http"
)

func HandleFlashers(cfg *config.Config, repo db.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		switch r.Method {

		case http.MethodGet:

			flashers, err := repo.ListFlashers()
			if err != nil {
				http.Error(
					w,
					err.Error(),
					http.StatusInternalServerError,
				)
				return
			}


			if flashers == nil {
				flashers = []db.FlasherRecord{}
			}


			w.Header().Set(
				"Content-Type",
				"application/json",
			)


			if err := json.NewEncoder(w).Encode(flashers); err != nil {
				http.Error(
					w,
					err.Error(),
					http.StatusInternalServerError,
				)
				return
			}


		default:

			http.Error(
				w,
				"method not allowed",
				http.StatusMethodNotAllowed,
			)
		}
	}
}