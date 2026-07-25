package handlers

import (
	"encoding/json"
	"flasher/config"
	"flasher/db" 
	"net/http" 
)

func HandleFirmwares(cfg *config.Config, repo db.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		switch r.Method {

		case http.MethodGet:

			firmwares, err := repo.ListFirmwares()
			if err != nil {
				http.Error(
					w,
					err.Error(),
					http.StatusInternalServerError,
				)
				return
			}

			if firmwares == nil {
				firmwares = []db.FirmwareRecord{}
			}

			w.Header().Set(
				"Content-Type",
				"application/json",
			)

			if err := json.NewEncoder(w).Encode(firmwares); err != nil {
				http.Error(
					w,
					err.Error(),
					http.StatusInternalServerError,
				)
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