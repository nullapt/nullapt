package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/nullapt/nullapt/cmd/registry-api/handlers"
	"github.com/nullapt/nullapt/internal/blob"
	"github.com/nullapt/nullapt/internal/db"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is required")
	}
	blobToken := os.Getenv("BLOB_READ_WRITE_TOKEN")
	if blobToken == "" {
		log.Fatal("BLOB_READ_WRITE_TOKEN is required")
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := db.Connect(ctx, dsn)
	if err != nil {
		log.Fatalf("connecting to database: %v", err)
	}
	defer pool.Close()

	blobClient := blob.New(blobToken)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)
	r.Use(corsMiddleware)

	skills := handlers.NewSkillsHandler(pool, blobClient)
	authMw := handlers.AuthMiddleware(pool)
	optionalAuth := handlers.OptionalAuth(pool)

	r.Route("/v1", func(r chi.Router) {
		r.With(optionalAuth).Get("/skills", skills.List)
		r.With(optionalAuth).Get("/skills/{name}", skills.Get)
		r.With(optionalAuth).Get("/skills/{name}/{version}", skills.Get)
		r.Get("/skills/{name}/log", skills.TransparencyLog)
		r.With(authMw).Post("/skills", skills.Publish)
	})

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "ok")
	})

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 60 * time.Second, // longer for WASM uploads
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("registry-api listening on :%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down...")
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(ctx) //nolint:errcheck
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
