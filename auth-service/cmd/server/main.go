package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"bookshelf/auth-service/internal/config"
	"bookshelf/auth-service/internal/handler"
	internalMiddleware "bookshelf/auth-service/internal/middleware"
	"bookshelf/auth-service/internal/repository"
	"bookshelf/auth-service/internal/service"

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

	// repository
	userRepo := repository.NewUserRepository(db)

	// service
	userService := service.NewUserService(
		userRepo,
		cfg.JWT.Secret,
	)

	// handler
	authHandler := handler.NewAuthHandler(userService)
	internalHandler := handler.NewInternalHandler(userService)
	healthHandler := handler.NewHealthHandler(db, cfg.App.Name, cfg.App.Version)

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
			"X-Service-Key",
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

	// health and ready routes
	r.Get("/health", healthHandler.Health)


	// internal routes
	r.Route("/internal/v1", func(r chi.Router) {
		r.Use(internalMiddleware.ServiceKeyMiddleware(cfg.Internal.ServiceKey))

		r.Route("/auth", func(r chi.Router) {
			r.Post("/verify", internalHandler.VerifyToken)
		})

		r.Post("/users/batch", internalHandler.GetUsersByIDs)
	})

	// public routes
	r.Route("/api/v1/auth", func(r chi.Router) {
		r.Get("/ready", handler.ReadyHandler(db))
		r.Post("/register", authHandler.Register)
		r.Post("/login", authHandler.Login)
	})

	// protected routes
	r.Group(func(r chi.Router) {
		r.Use(authHandler.AuthMiddleware)

		r.Get("/api/v1/users/me", authHandler.GetMe)
		r.Put("/api/v1/users/me", authHandler.UpdateMe)
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
