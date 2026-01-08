// Package server provides HTTP API capabilities
package server

import (
	"net/http"
	"time"

	"github.com/AkMo3/simplify/internal/store"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	router *chi.Mux
	store  *store.Store
}

// New only requires Store now, Logger is global
func New(storeImpl *store.Store) *Server {
	s := &Server{
		router: chi.NewRouter(),
		store:  storeImpl,
	}
	s.setupMiddleware()
	s.setupRoutes()
	return s
}

func (s *Server) setupMiddleware() {
	s.router.Use(middleware.RequestID)
	s.router.Use(middleware.RealIP)
	s.router.Use(middleware.Logger)
	s.router.Use(middleware.Recoverer)
	s.router.Use(middleware.AllowContentType("application/json"))
}

func (s *Server) setupRoutes() {
	s.router.Route("/api/v1", func(r chi.Router) {
		r.Post("/applications", s.handleCreateApplication)
		r.Get("/applications", s.handleListApplications)
		r.Get("/applications/{id}", s.handleGetApplication)
		r.Delete("/applications/{id}", s.handleDeleteApplication)
	})
}

func (s *Server) Start(addr string) error {
	// Using global logger directly if needed, or just standard log
	server := http.Server{
		Addr:        addr,
		ReadTimeout: 2 * time.Second,
		Handler:     s.router,
	}
	return server.ListenAndServe()
}
