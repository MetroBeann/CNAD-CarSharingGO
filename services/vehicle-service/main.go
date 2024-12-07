// Path: services/vehicle-service/main.go
package main

import (
    "database/sql"
    "fmt"
    "log"
    "net/http"
    "os"
    "time"
    "path/filepath"

    gorillaCORS "github.com/gorilla/handlers"
    "github.com/gorilla/mux"
    _ "github.com/lib/pq"

    "vehicle-service/handlers"
    "vehicle-service/repository"
    "vehicle-service/middleware"
)

const (
    defaultPort = ":8081"  // Different port from user-service
    dbConnRetries = 5
    dbConnRetryDelay = 5 * time.Second
)

func initDB(connStr string) (*sql.DB, error) {
    var db *sql.DB
    var err error

    for i := 0; i < dbConnRetries; i++ {
        db, err = sql.Open("postgres", connStr)
        if err != nil {
            log.Printf("Failed to open database connection (attempt %d/%d): %v", i+1, dbConnRetries, err)
            time.Sleep(dbConnRetryDelay)
            continue
        }

        err = db.Ping()
        if err == nil {
            log.Println("Successfully connected to the database")
            break
        }

        log.Printf("Failed to ping database (attempt %d/%d): %v", i+1, dbConnRetries, err)
        db.Close()
        time.Sleep(dbConnRetryDelay)
    }

    if err != nil {
        return nil, fmt.Errorf("failed to establish database connection after %d attempts: %v", dbConnRetries, err)
    }

    db.SetMaxOpenConns(25)
    db.SetMaxIdleConns(25)
    db.SetConnMaxLifetime(5 * time.Minute)

    return db, nil
}

func setupRoutes(vehicleHandler *handlers.VehicleHandler) *mux.Router {
    r := mux.NewRouter()

    // API routes
    api := r.PathPrefix("/api").Subrouter()
    
    // Vehicle routes
    api.HandleFunc("/vehicles/available", middleware.AuthMiddleware(vehicleHandler.GetAvailableVehicles)).Methods("POST", "OPTIONS")
    api.HandleFunc("/reservations", middleware.AuthMiddleware(vehicleHandler.CreateReservation)).Methods("POST", "OPTIONS")
    api.HandleFunc("/reservations/user", middleware.AuthMiddleware(vehicleHandler.GetUserReservations)).Methods("GET", "OPTIONS")
    api.HandleFunc("/reservations/{id}", middleware.AuthMiddleware(vehicleHandler.UpdateReservation)).Methods("PUT", "OPTIONS")
    api.HandleFunc("/reservations/{id}", middleware.AuthMiddleware(vehicleHandler.CancelReservation)).Methods("DELETE", "OPTIONS")

	api.HandleFunc("/verify-token", middleware.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).Methods("GET", "OPTIONS")
	
    // Static file server for the frontend directory
    fs := http.FileServer(http.Dir("frontend"))
    
    // Handle the root path
    r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        http.ServeFile(w, r, "frontend/index.html")
    })
    
    // Serve other static files
    r.PathPrefix("/").Handler(fs)

    return r
}

func setupCORS(handler http.Handler) http.Handler {
    return gorillaCORS.CORS(
        gorillaCORS.AllowedOrigins([]string{"http://localhost:8080", "http://localhost:8081"}),
        gorillaCORS.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
        gorillaCORS.AllowedHeaders([]string{"Content-Type", "Authorization"}),
        gorillaCORS.AllowCredentials(),
        gorillaCORS.MaxAge(86400),
    )(handler)
}

func main() {
    log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
    
    // Database connection string (use same Supabase instance)
    connStr := "postgres://postgres.wjdhhzmaclmsvaiszagk:22KC6282t04@@aws-0-ap-southeast-1.pooler.supabase.com:6543/postgres"
    
    db, err := initDB(connStr)
    if err != nil {
        log.Fatal("Failed to initialize database:", err)
    }
    defer db.Close()

    // Initialize repositories and handlers
    vehicleRepo := repository.NewVehicleRepository(db)
    reservationRepo := repository.NewReservationRepository(db)
    vehicleHandler := handlers.NewVehicleHandler(vehicleRepo, reservationRepo)

    // Setup routes
    router := setupRoutes(vehicleHandler)

    // Setup CORS
    corsHandler := setupCORS(router)
	

    // Determine port
    port := os.Getenv("PORT")
    if port == "" {
        port = defaultPort
    }

    // Create server
    server := &http.Server{
        Addr:         port,
        Handler:      corsHandler,
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 15 * time.Second,
        IdleTimeout:  60 * time.Second,
    }

    // Start server
    log.Printf("Vehicle Service starting on port %s\n", port)
    log.Printf("Serving frontend from: %s\n", filepath.Join(".", "frontend"))
    if err := server.ListenAndServe(); err != nil {
        log.Fatal("Server failed to start:", err)
    }
}