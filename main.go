package main

import (
	// "crypto/aes"
	// "crypto/cipher"
	//"crypto/hmac"
	// "crypto/rand"
	// "crypto/sha256"
	// "encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
  	"html/template"
	"os"
	"path/filepath"
	"strings"
	"time" 
  	"flasher/config" 
  	"flasher/handlers"
	"flasher/db" 
)
 
  func main() {
	fmt.Println("SERVER STARTED FROM CURRENT SOURCE")

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
 

	repo, err := db.NewSQLiteRepository("./database.sqlite3")
	if err != nil {
		log.Fatal(err)
	}

	defer repo.Close()

	if err := os.MkdirAll(cfg.FirmwaresDir, 0755); err != nil {
		log.Fatal(err)
	}

	base := template.Must(template.ParseFiles(
		"ui/templates/layout.html",
		"ui/templates/header.html",
		"ui/templates/download_flasher_btn.html",
	))

	businessTmpl := template.Must(base.Clone())
	template.Must(businessTmpl.ParseFiles("ui/pages/business.html"))

	techTmpl := template.Must(base.Clone())
	template.Must(techTmpl.ParseFiles("ui/pages/tech.html"))

	adminTmpl := template.Must(base.Clone())
	template.Must(adminTmpl.ParseFiles("ui/pages/admin.html"))

	mux := http.NewServeMux()

	mux.Handle("/css/",
		http.StripPrefix("/css/",
			http.FileServer(http.Dir("./ui/css")),
		),
	)

	mux.Handle("/js/",
		http.StripPrefix("/js/",
			http.FileServer(http.Dir("./ui/js")),
		),
	)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		if err := businessTmpl.ExecuteTemplate(w, "layout", nil); err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	mux.HandleFunc("/tech", func(w http.ResponseWriter, r *http.Request) {
		if err := techTmpl.ExecuteTemplate(w, "layout", nil); err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	mux.HandleFunc("/admin", func(w http.ResponseWriter, r *http.Request) {
		if err := adminTmpl.ExecuteTemplate(w, "layout", nil); err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	mux.HandleFunc("/list_firmwares", handleListFirmwares(cfg, repo))
	mux.HandleFunc("/download_firmware/", handleDownloadFirmware(cfg, repo))
	mux.HandleFunc("/upload_firmware", handleUploadFirmware(cfg, repo))
	mux.HandleFunc("/delete_firmware/", handleDeleteFirmware(cfg, repo))

	mux.HandleFunc("/list_flashers", handleListFlashers(cfg, repo))
	mux.HandleFunc("/download_flasher/", handleDownloadFlasher(cfg, repo))
	mux.HandleFunc("/upload_flasher", handleUploadFlasher(cfg, repo))
	mux.HandleFunc("/delete_flasher/", handleDeleteFlasher(cfg, repo))

	fmt.Printf("IgnitionFlash Admin running on %s\n", cfg.ListenAddr)

	if err := http.ListenAndServe(cfg.ListenAddr, mux); err != nil {
	log.Println(err)
}
 	 
}
 