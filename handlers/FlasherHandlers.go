package handlers

import (
	"flasher/db" 
	"net/http"
	"flasher/config" 
	"log"
)

func handleUploadFlasher (cfg *config.Config, database *db.Repository) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    // Implementation for uploading the flasher

    log.Printf("")
  }
}


func handleDownloadFlasher (cfg *config.Config, database *db.Repository) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    // Implementation for downloading the flasher
  }
}

func  handleListFlashers (cfg *config.Config, database *db.Repository) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    // Implementation for listing flashers
  }
}

func handleDeleteFlasher (cfg *config.Config, database *db.Repository) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    // Implementation for deleting a flasher
  }
 }