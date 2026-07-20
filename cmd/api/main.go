package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kevstaa/taskflow/config"
	"github.com/kevstaa/taskflow/db"
	"github.com/kevstaa/taskflow/internal/handler"
	"github.com/kevstaa/taskflow/internal/middleware"
	"github.com/kevstaa/taskflow/internal/repository"
	"github.com/kevstaa/taskflow/internal/service"
	"github.com/redis/go-redis/v9"
)

func main() {
	cfg := config.Load()

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Connection error:", err)
	}
	defer pool.Close()

	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	defer rdb.Close()

	queries := db.New(pool)

	// repositories
	userRepo := repository.NewUserRepository(queries)
	projectRepo := repository.NewProjectRepository(queries)
	taskRepo := repository.NewTaskRepository(queries)
	memberRepo := repository.NewProjectMemberRepository(queries)

	// services
	authService := service.NewAuthService(userRepo, cfg.JWTSecret)
	userService := service.NewUserService(userRepo)
	projectService := service.NewProjectService(projectRepo)
	taskService := service.NewTaskService(taskRepo)
	memberService := service.NewProjectMemberService(memberRepo)

	// middleware
	authMiddleware := middleware.NewAuthMiddleware(cfg.JWTSecret)

	// handlers
	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(userService)
	projectHandler := handler.NewProjectHandler(projectService)
	taskHandler := handler.NewTaskHandler(taskService)
	memberHandler := handler.NewProjectMemberHandler(memberService)

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Post("/auth/register", authHandler.Register)
	r.Post("/auth/login", authHandler.Login)

	r.Group(func(r chi.Router) {
		r.Use(authMiddleware.Authenticate)

		r.Get("/api/users/me", userHandler.GetMe)
		r.Put("/api/users/me", userHandler.UpdateMe)

		r.Route("/api/projects", func(r chi.Router) {
			r.Get("/", projectHandler.GetByUser)
			r.Post("/", projectHandler.Create)
			r.Get("/{id}", projectHandler.GetByID)
			r.Put("/{id}", projectHandler.Update)
			r.Delete("/{id}", projectHandler.Delete)

			r.Get("/{id}/members", memberHandler.GetProjectMembers)
			r.Post("/{id}/members", memberHandler.AddMember)
			r.Delete("/{id}/members/{userId}", memberHandler.RemoveMember)
		})

		r.Route("/api/projects/{id}/tasks", func(r chi.Router) {
			r.Get("/", taskHandler.GetByProject)
			r.Post("/", taskHandler.Create)
			r.Get("/{tid}", taskHandler.GetByID)
			r.Put("/{tid}", taskHandler.Update)
			r.Delete("/{tid}", taskHandler.Delete)
		})
	})

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Println("Server listening on " + cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Error starting server:", err)
		}
	}()

	<-quit
	log.Println("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatal("Error shutting down:", err)
	}

	log.Println("Server stopped")
}
