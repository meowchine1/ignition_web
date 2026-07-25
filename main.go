package main

import (
	// "crypto/aes"
	// "crypto/cipher"
	//"crypto/hmac"
	// "crypto/rand"
	// "crypto/sha256"
	// "encoding/binary"
	//"encoding/json"
	"fmt"
	//"io"
	"log"
	"net/http"
  	"html/template"
	"os"
	//"path/filepath"
	//"strings"
	//"time" 
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
 
	mux.HandleFunc("/api/firmwares", handlers.HandleFirmwares(cfg, repo))  // GET
	mux.HandleFunc("/api/firmware", handlers.HandleFirmware(cfg, repo))	   // POST
	mux.HandleFunc("/api/firmware/", handlers.HandleFirmware(cfg, repo))   // GET DELETE

	mux.HandleFunc("/api/flashers", handlers.HandleFlashers(cfg, repo))   // GET
	mux.HandleFunc("/api/flasher", handlers.HandleFlasher(cfg, repo))	  // POST
	mux.HandleFunc("/api/flasher/", handlers.HandleFlasher(cfg, repo))	  // GET DELETE
	mux.HandleFunc("/api/flasher/current", handlers.HandleCurrentFlasher(cfg, repo)) // GET

	fmt.Printf("IgnitionFlash Admin running on %s\n", cfg.ListenAddr)

	if err := http.ListenAndServe(cfg.ListenAddr, mux); err != nil {
	log.Println(err)
}
 	 
}
 