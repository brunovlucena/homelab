package router

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/rs/cors"
)

// Struct which defines a customized ChiRouter.
type MyRouter struct {
	Mux *chi.Mux
}

// NewRouter returns a MyRouter
func NewRouter() *MyRouter {
	r := &MyRouter{Mux: chi.NewRouter()}
	// Cors
	cors := setupCors()
	// Begin Middleware
	r.Mux.Use(middleware.Logger)
	r.Mux.Use(middleware.URLFormat)
	r.Mux.Use(middleware.Compress(5, "gzip"))
	r.Mux.Use(cors.Handler)
	r.Mux.Use(render.SetContentType(render.ContentTypeJSON))
	// End Middleware
	return r
}

// SetupCors configures CORS net/http middleware
func setupCors() *cors.Cors {
	cors := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	})
	return cors
}
