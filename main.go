package main

import (
	"flag"
	"flasher/config"
	"flasher/db"
	"flasher/handlers"
	"flasher/middleware"
	"flasher/services"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
)

func main() {
	fmt.Println("SERVER STARTED FROM CURRENT SOURCE")

	// Разовый bootstrap админа: go run . -create-admin -admin-user=admin -admin-pass=...
	createAdmin := flag.Bool("create-admin", false, "create an admin user and exit")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	repo, err := db.NewSQLiteRepository("./database.sqlite3")
	if err != nil {
		log.Fatal(err)
	}
	defer repo.Close()

	authService := services.NewAuthService(cfg, repo)

	if *createAdmin {
		if cfg.AdminUsername == "" || cfg.AdminPassword == "" {
			log.Fatal("ADMIN_USERNAME and ADMIN_PASSWORD env vars are required")
		}
		if _, err := authService.CreateAdmin(cfg.AdminUsername, cfg.AdminPassword); err != nil {
			log.Fatalf("failed to create admin: %v", err)
		}
		fmt.Println("admin created")
		return
	}

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

	mux.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("./ui/css"))))
	mux.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("./ui/js"))))

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

	requireAuth := middleware.RequireAuth(authService)
	requireAdmin := middleware.RequireRole(db.RoleAdmin)

	// ---------------- Auth (публичные) ----------------

	authHandler := handlers.NewAuthHandler(authService)

	mux.HandleFunc("POST /api/auth/register", authHandler.Register)
	mux.HandleFunc("POST /api/auth/login", authHandler.Login)
	mux.HandleFunc("POST /api/auth/logout", middleware.Chain(authHandler.Logout, requireAuth))

	// ---------------- Firmware ----------------

	firmwareService := services.NewFirmwareService(cfg, repo)
	firmwareHandler := handlers.NewFirmwareHandler(firmwareService)

	// Доступно user и admin
	mux.HandleFunc("GET /api/firmwares/available", middleware.Chain(firmwareHandler.ListAvailable, requireAuth))
	mux.HandleFunc("GET /api/firmwares/available/{id}", middleware.Chain(firmwareHandler.GetAvailable, requireAuth))

	// Только admin
	mux.HandleFunc("GET /api/firmwares", middleware.Chain(firmwareHandler.List, requireAuth, requireAdmin))
	mux.HandleFunc("POST /api/firmware", middleware.Chain(firmwareHandler.Create, requireAuth, requireAdmin))
	mux.HandleFunc("GET /api/firmware/{id}", middleware.Chain(firmwareHandler.Get, requireAuth, requireAdmin))
	mux.HandleFunc("PATCH /api/firmware/{id}", middleware.Chain(firmwareHandler.Update, requireAuth, requireAdmin))
	mux.HandleFunc("DELETE /api/firmware/{id}", middleware.Chain(firmwareHandler.Delete, requireAuth, requireAdmin))

	// ---------------- Flasher ----------------

	flasherService := services.NewFlasherService(cfg, repo)
	flasherHandler := handlers.NewFlasherHandler(flasherService)

	// Доступно user и admin
	mux.HandleFunc("GET /api/flashers/current", middleware.Chain(flasherHandler.Current, requireAuth))

	// Только admin
	mux.HandleFunc("GET /api/flashers", middleware.Chain(flasherHandler.List, requireAuth, requireAdmin))
	mux.HandleFunc("POST /api/flasher", middleware.Chain(flasherHandler.Create, requireAuth, requireAdmin))
	mux.HandleFunc("GET /api/flasher/{id}", middleware.Chain(flasherHandler.Get, requireAuth, requireAdmin))
	mux.HandleFunc("PATCH /api/flasher/{id}", middleware.Chain(flasherHandler.Update, requireAuth, requireAdmin))
	mux.HandleFunc("DELETE /api/flasher/{id}", middleware.Chain(flasherHandler.Delete, requireAuth, requireAdmin))

	fmt.Printf("IgnitionFlash Admin running on %s\n", cfg.ListenAddr)

	if err := http.ListenAndServe(cfg.ListenAddr, mux); err != nil {
		log.Println(err)
	}
}