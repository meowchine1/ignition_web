package main

import (
	"bufio"
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
	"strings"
	"syscall"
	"path/filepath"
	"golang.org/x/term"
)

func main() {
	fmt.Println("SERVER STARTED")

	// Первоначальная генерация machine secrets: go run . -init-secrets
	initSecrets := flag.Bool("init-secrets", false, "generate missing machine secrets in .env and exit")
	// Создание первого администратора: go run . -create-admin
	createAdmin := flag.Bool("create-admin", false, "interactively create the first admin user and exit")
	flag.Parse()

	// Генерация секретов — до любой загрузки конфигурации.
	if *initSecrets {
		changed, err := config.InitSecrets()
		if err != nil {
			log.Fatal(err)
		}
		if changed {
			fmt.Println("secrets generated/updated in .env")
		} else {
			fmt.Println("all secrets already present in .env — nothing to do")
		}
		return
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	dbPath := filepath.Join(cfg.DatabaseDir, config.DatabaseFileName)

	repo, err := db.NewSQLiteRepository(dbPath)
	if err != nil {
		log.Fatal(err)
	}
	 
	defer repo.Close()

	authService := services.NewAuthService(cfg, repo)

	// Создание первого администратора — локальная CLI-команда, не HTTP endpoint.
	if *createAdmin {
		if err := createFirstAdminCLI(authService); err != nil {
			log.Fatal(err)
		}
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

	loginTmpl := template.Must(base.Clone())
	template.Must(loginTmpl.ParseFiles("ui/pages/login.html"))

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

	mux.HandleFunc("/admin", middleware.RequireAdminPage(authService)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := adminTmpl.ExecuteTemplate(w, "layout", nil); err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})).ServeHTTP)

	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if err := loginTmpl.ExecuteTemplate(w, "layout", nil); err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	requireAuth := middleware.RequireAuth(authService)
	requireAdmin := middleware.RequireRole(db.RoleAdmin)

	// ---------------- Auth (публичные) ----------------

	authHandler := handlers.NewAuthHandler(authService)
	usersHandler := handlers.NewUsersHandler(authService)

	mux.HandleFunc("POST /api/auth/register", authHandler.Register)
	mux.HandleFunc("POST /api/auth/login", authHandler.Login)
	mux.HandleFunc("POST /api/auth/logout", middleware.Chain(authHandler.Logout, requireAuth))
	mux.HandleFunc("POST /api/auth/change-password", middleware.Chain(usersHandler.ChangePassword, requireAuth))

	// ---------------- Users (admin only) ----------------

	mux.HandleFunc("GET /api/users", middleware.Chain(usersHandler.List, requireAuth, requireAdmin))
	mux.HandleFunc("POST /api/users", middleware.Chain(usersHandler.Create, requireAuth, requireAdmin))
	mux.HandleFunc("PATCH /api/users/{id}", middleware.Chain(usersHandler.Update, requireAuth, requireAdmin))
	mux.HandleFunc("DELETE /api/users/{id}", middleware.Chain(usersHandler.Delete, requireAuth, requireAdmin))
	mux.HandleFunc("POST /api/users/{id}/reset-password", middleware.Chain(usersHandler.ResetPassword, requireAuth, requireAdmin))

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

// createFirstAdminCLI интерактивно запрашивает username/password и создаёт первого admin.
func createFirstAdminCLI(authService *services.AuthService) error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter admin username: ")
	username, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read username: %w", err)
	}
	username = strings.TrimSpace(username)
	if username == "" {
		return fmt.Errorf("username cannot be empty")
	}

	fmt.Print("Enter admin password: ")
	password, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return fmt.Errorf("failed to read password: %w", err)
	}
	fmt.Println()

	fmt.Print("Confirm admin password: ")
	confirm, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return fmt.Errorf("failed to read password confirmation: %w", err)
	}
	fmt.Println()

	if string(password) != string(confirm) {
		return fmt.Errorf("passwords do not match")
	}

	user, err := authService.CreateFirstAdmin(username, string(password))
	if err != nil {
		if err == services.ErrAdminExists {
			return fmt.Errorf("admin already exists — initial setup is not allowed to overwrite it")
		}
		return err
	}

	fmt.Printf("Admin user %q created successfully (ID=%d)\n", user.Username, user.ID)
	return nil
}