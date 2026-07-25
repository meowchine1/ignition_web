package handlers

import (
	"encoding/json"
	"flasher/config"
	"flasher/db"
	"net/http"
	"path"
	"strconv"
)

func HandleFlasher(cfg *config.Config, repo db.Repository) http.HandlerFunc {

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

			// if !checkAdmin(r, cfg) {
			// 	http.Error(w, "unauthorized", http.StatusUnauthorized)
			// 	return
			// }


			// TODO:
			// r.ParseMultipartForm()
			// сохранить файл
			// создать db.FlasherRecord
			// repo.AddFlasher()


			w.WriteHeader(http.StatusCreated)


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


			// if !checkAdmin(r, cfg) {
			// 	http.Error(w, "unauthorized", http.StatusUnauthorized)
			// 	return
			// }


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