package main

import (
	"bookshelf/books-service/internal/client"
	"bookshelf/books-service/internal/config"
	"bookshelf/books-service/internal/handler"
	authMiddleware "bookshelf/books-service/internal/middleware"
	"bookshelf/books-service/internal/repository"
	"bookshelf/books-service/internal/service"
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	chiMiddleware "github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func main() {
	// config
	cfg := config.Load()

	// DB
	db, err := sqlx.Connect("postgres", cfg.Database.URL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)
	log.Println("Connected to database")

	// repositories
	bookRepo := repository.NewBookRepository(db)
	reviewRepo := repository.NewReviewRepository(db)

	// services
	bookService := service.NewBookService(bookRepo)
	reviewService := service.NewReviewService(reviewRepo, bookRepo)

	// handlers
	bookHandler := handler.NewBookHandler(bookService)
	reviewHandler := handler.NewReviewHandler(reviewService)

	// client
	authClient := client.NewAuthClient(cfg.AuthService.URL, cfg.AuthService.Timeout, cfg.AuthService.ServiceKey)

	// router
	r := chi.NewRouter()

	// CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{
			"http://localhost:5174",
		},
		AllowedMethods: []string{
			"GET",
			"POST",
			"PUT",
			"DELETE",
			"OPTIONS",
		},
		AllowedHeaders: []string{
			"Accept",
			"Authorization",
			"Content-Type",
		},
		ExposedHeaders: []string{
			"Link",
		},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// global middleware
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", handler.HealthHandler)
		r.Get("/ready", handler.ReadyHandler(db))

		// ===== Public =====

		r.Get("/books", bookHandler.List)
		r.Get("/books/{id}", bookHandler.GetByID)
		r.Get("/books/{book_id}/reviews", reviewHandler.List)

		// ===== Protected =====

		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.AuthMiddleware(authClient))

			r.Post("/books", bookHandler.Create)
			r.Put("/books/{id}", bookHandler.Update)
			r.Delete("/books/{id}", bookHandler.Delete)

			r.Post("/books/{book_id}/reviews", reviewHandler.Create)

			r.Put("/reviews/{id}", reviewHandler.Update)
			r.Delete("/reviews/{id}", reviewHandler.Delete)
		})
	})

	// graceful shutdown
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	go func() {
		log.Printf("Server started on %s", srv.Addr)

		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server failed: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("server shutdown failed: %v", err)
	}

	log.Println("Server stopped")
}
