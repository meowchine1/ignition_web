package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"path/filepath"
	"strconv"

	"flasher/services"
)

type FlasherHandler struct {
	service *services.FlasherService
}

func NewFlasherHandler(service *services.FlasherService) *FlasherHandler {
	return &FlasherHandler{service: service}
}

func flasherPathID(r *http.Request) (int64, error) {
	return strconv.ParseInt(r.PathValue("id"), 10, 64)
}

// GET /api/flashers?os=...  — admin only, полный список (с фильтром по os)
func (h *FlasherHandler) List(w http.ResponseWriter, r *http.Request) { 
	osName := r.PathValue("os")

    if osName == "" {
        http.Error(w, "os is required", http.StatusBadRequest)
        return
    }
	
	flashers, err := h.service.List(osName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, flashers)
}

// GET /api/flashers/current?os=...  — доступно роли user (и admin)
func (h *FlasherHandler) Current(w http.ResponseWriter, r *http.Request) { 
	osName := r.PathValue("os")

    if osName == "" {
        http.Error(w, "os is required", http.StatusBadRequest)
        return
    }

	flasher, err := h.service.Current(osName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	http.ServeFile(w, r, flasher.Path)
}

// GET /api/flasher/{id} — скачивание файла флешера
func (h *FlasherHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := flasherPathID(r)
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}

	flasher, err := h.service.Get(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	http.ServeFile(w, r, flasher.Path)
}

// POST /api/flasher
func (h *FlasherHandler) Create(w http.ResponseWriter, r *http.Request) {
	const maxUploadSize = 100 << 20 // 100 MB

	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		http.Error(w, "invalid multipart form", http.StatusBadRequest)
		return
	}

	osType := r.FormValue("os")
	if osType == "" {
		http.Error(w, "os is required", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("flasher")
	if err != nil {
		http.Error(w, "flasher file is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	raw, err := io.ReadAll(file)
	if err != nil || len(raw) == 0 {
		http.Error(w, "empty file", http.StatusBadRequest)
		return
	}

	record, err := h.service.Create(osType, header.Filename, raw)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"ok":   true,
		"file": filepath.Base(record.Path),
		"path": record.Path,
	})
}

// PATCH /api/flasher/{id}
func (h *FlasherHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := flasherPathID(r)
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}

	var request struct {
		OS      string `json:"os"`
		Current bool   `json:"current"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if err := h.service.SetCurrent(id, request.OS, request.Current); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DELETE /api/flasher/{id}
func (h *FlasherHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := flasherPathID(r)
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}

	if err := h.service.Delete(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}