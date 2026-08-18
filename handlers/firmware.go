package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"flasher/services"
)

type FirmwareHandler struct {
	service *services.FirmwareService
}

func NewFirmwareHandler(service *services.FirmwareService) *FirmwareHandler {
	return &FirmwareHandler{service: service}
}

func pathID(r *http.Request) (int, error) {
	return strconv.Atoi(r.PathValue("id"))
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// GET /api/firmwares
func (h *FirmwareHandler) List(w http.ResponseWriter, r *http.Request) {
	firmwares, err := h.service.List()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, firmwares)
}

// GET /api/firmwares/available
func (h *FirmwareHandler) ListAvailable(w http.ResponseWriter, r *http.Request) {
	firmwares, err := h.service.ListAvailable()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, firmwares)
}

// GET /api/firmware/{id} — скачивание файла прошивки
func (h *FirmwareHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}

	firmware, err := h.service.Get(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	http.ServeFile(w, r, firmware.Path)
}

// GET /api/firmwares/available/{id} — скачивание доступной прошивки
func (h *FirmwareHandler) GetAvailable(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}

	firmware, err := h.service.GetAvailable(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	http.ServeFile(w, r, firmware.Path)
}

// POST /api/firmware
func (h *FirmwareHandler) Create(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("firmware")
	if err != nil {
		http.Error(w, "no file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	if !strings.HasSuffix(header.Filename, ".bin") {
		http.Error(w, "only .bin allowed", http.StatusBadRequest)
		return
	}

	raw, err := io.ReadAll(file)
	if err != nil || len(raw) == 0 {
		http.Error(w, "empty file", http.StatusBadRequest)
		return
	}

	record, err := h.service.Create(filepath.Base(header.Filename), raw)
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

// PATCH /api/firmware/{id}
func (h *FirmwareHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}

	var request struct {
		Available bool `json:"available"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if err := h.service.SetAvailable(id, request.Available); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DELETE /api/firmware/{id}
func (h *FirmwareHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
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