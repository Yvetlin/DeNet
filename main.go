package main

import (
	"log"
	"net/http"
	"os"

	"DeNet/database"
	"DeNet/handlers"
	"DeNet/middleware"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	if err := database.Init(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	r := mux.NewRouter()
	
	r.Use(middleware.RecoveryMiddleware)

	api := r.PathPrefix("/users").Subrouter()
	api.Use(middleware.AuthMiddleware)

	api.HandleFunc("/{id}/status", handlers.GetUserStatus).Methods("GET")
	api.HandleFunc("/leaderboard", handlers.GetLeaderboard).Methods("GET")
	api.HandleFunc("/{id}/task/complete", handlers.CompleteTask).Methods("POST")
	api.HandleFunc("/{id}/referrer", handlers.SetReferrer).Methods("POST")

	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	// Эндпоинт для генерации токена (без авторизации, только для тестирования)
	r.HandleFunc("/auth/token/{id}", handlers.GenerateToken).Methods("GET")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
