package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"taskflow/internal/config"
	"taskflow/internal/database"
	"taskflow/internal/handler"
	appmiddleware "taskflow/internal/middleware"
	"taskflow/internal/repository"
	"taskflow/internal/service"
	"taskflow/internal/ws"
)

func main() {
	cfg := config.Load()

	db, err := database.New(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	healthHandler := handler.NewHealthHandler(db)

	userRepo := repository.NewUserRepository(db)
	authService := service.NewAuthService(userRepo, cfg.JWTSecret)
	authHandler := handler.NewAuthHandler(authService, userRepo)

	boardRepo := repository.NewBoardRepository(db)
	listRepo := repository.NewListRepository(db)
	cardRepo := repository.NewCardRepository(db)

	hub := ws.NewHub()

	boardService := service.NewBoardService(boardRepo, listRepo, cardRepo, hub)
	listService := service.NewListService(boardRepo, listRepo, db, hub)
	cardService := service.NewCardService(boardRepo, listRepo, cardRepo, db, hub)

	boardHandler := handler.NewBoardHandler(boardService)
	listHandler := handler.NewListHandler(listService)
	cardHandler := handler.NewCardHandler(cardService)
	wsHandler := handler.NewWSHandler(hub, boardService, authService)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
	}))

	r.Get("/health", healthHandler.Check)

	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", authHandler.Register)
		r.Post("/login", authHandler.Login)

		r.Group(func(r chi.Router) {
			r.Use(appmiddleware.Auth(authService))
			r.Get("/me", authHandler.Me)
		})
	})

	r.Route("/boards", func(r chi.Router) {
		r.Use(appmiddleware.Auth(authService))

		r.Post("/", boardHandler.Create)
		r.Get("/", boardHandler.List)

		r.Route("/{boardID}", func(r chi.Router) {
			r.Get("/", boardHandler.GetDetail)
			r.Put("/", boardHandler.Update)
			r.Delete("/", boardHandler.Delete)

			r.Post("/lists", listHandler.Create)

			r.Route("/lists/{listID}", func(r chi.Router) {
				r.Put("/", listHandler.Update)
				r.Patch("/reorder", listHandler.Reorder)
				r.Delete("/", listHandler.Delete)

				r.Post("/cards", cardHandler.Create)

				r.Route("/cards/{cardID}", func(r chi.Router) {
					r.Put("/", cardHandler.Update)
					r.Patch("/move", cardHandler.Move)
					r.Delete("/", cardHandler.Delete)
				})
			})
		})
	})

	// Auth for this route is verified inside the handler via a "token" query
	// param, since browsers can't set custom headers on WebSocket handshakes.
	r.Get("/ws/boards/{boardID}", wsHandler.Serve)

	log.Printf("TaskFlow API listening on :%s (env=%s)", cfg.Port, cfg.Environment)
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
