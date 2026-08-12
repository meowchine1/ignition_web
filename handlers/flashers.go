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

				osStr := r.URL.Query().Get("os")

				var (
					flashers []db.FlasherRecord
					err      error
				)

				if osStr == "" {

					// Полный список
					flashers, err = repo.ListFlashers()

				} else {

					// Список для конкретной ОС
					flashers, err = repo.ListFlashersByOS(
						db.OSType(osStr),
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