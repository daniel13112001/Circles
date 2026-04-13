package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	firebase "firebase.google.com/go/v4"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"

	circlesauth "github.com/danielyakubu/circles/internal/auth"
	"github.com/danielyakubu/circles/internal/contacts"
	"github.com/danielyakubu/circles/internal/db"
	"github.com/danielyakubu/circles/internal/friends"
	"github.com/danielyakubu/circles/internal/groups"
	"github.com/danielyakubu/circles/internal/posts"
	"github.com/danielyakubu/circles/internal/users"
)

func main() {
	// Load .env if present (dev convenience; no-op in prod).
	_ = godotenv.Load()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	ctx := context.Background()

	// ── Database ─────────────────────────────────────────────────────────────
	pool, err := db.New(ctx)
	if err != nil {
		slog.Error("database init failed", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := db.Migrate(ctx, pool); err != nil {
		slog.Error("migrations failed", "err", err)
		os.Exit(1)
	}
	slog.Info("migrations applied")

	// ── Firebase ──────────────────────────────────────────────────────────────
	firebaseApp, err := initFirebase(ctx)
	if err != nil {
		slog.Error("firebase init failed", "err", err)
		os.Exit(1)
	}
	authClient, err := firebaseApp.Auth(ctx)
	if err != nil {
		slog.Error("firebase auth client failed", "err", err)
		os.Exit(1)
	}

	// ── Router ────────────────────────────────────────────────────────────────
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	// Health check (no auth).
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"status":"ok"}`)
	})

	// All routes below require a valid Firebase JWT.
	userSvc := users.NewService(users.NewRepo(pool))

	r.Group(func(r chi.Router) {
		r.Use(circlesauth.Middleware(authClient))

		// /users — POST does not need a resolved user (registering for first time)
		r.Mount("/users", users.Routes(userSvc))

		// Routes below also need a resolved internal user record.
		r.Group(func(r chi.Router) {
			r.Use(users.RequireUser(userSvc))

			friendsSvc := friends.NewService(friends.NewRepo(pool))
			postsSvc := posts.NewService(posts.NewRepo(pool))

			r.Mount("/contacts", contacts.Routes(contacts.NewService(contacts.NewRepo(pool))))
			r.Mount("/friends", friends.Routes(friendsSvc))
			r.Mount("/groups", groups.Routes(groups.NewService(groups.NewRepo(pool), friendsSvc)))
			r.Mount("/groups/{id}/posts", posts.GroupRoutes(postsSvc))
			r.Get("/feed", posts.GlobalFeedHandler(postsSvc))
		})
	})

	// ── Server ────────────────────────────────────────────────────────────────
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("server starting", "port", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	// Graceful shutdown on SIGINT / SIGTERM.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("shutdown error", "err", err)
	}
}

func initFirebase(ctx context.Context) (*firebase.App, error) {
	projectID := os.Getenv("FIREBASE_PROJECT_ID")
	if projectID == "" {
		return nil, fmt.Errorf("FIREBASE_PROJECT_ID is not set")
	}

	cfg := &firebase.Config{ProjectID: projectID}

	keyFile := os.Getenv("FIREBASE_SERVICE_ACCOUNT_FILE")
	if keyFile != "" {
		return firebase.NewApp(ctx, cfg, option.WithCredentialsFile(keyFile))
	}
	return firebase.NewApp(ctx, cfg)
}

