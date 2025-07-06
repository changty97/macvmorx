package web

import (
	"log"
	"net/http"
	"time"

	"github.com/changty97/macvmorx/internal/api"
	"github.com/changty97/macvmorx/internal/config"
	"github.com/gorilla/mux" // Using gorilla/mux for more advanced routing
)

// Server represents the web server for macvmorx.
type Server struct {
	Port    string
	Handler *api.Handlers
	Cfg     *config.Config // Still need config for port
}

// NewServer creates a new web server instance.
func NewServer(port string, handler *api.Handlers, cfg *config.Config) *Server {
	return &Server{
		Port:    port,
		Handler: handler,
		Cfg:     cfg,
	}
}

// Start runs the HTTP server.
func (s *Server) Start() {
	router := mux.NewRouter()

	// API routes
	router.HandleFunc("/api/heartbeat", s.Handler.HandleHeartbeat).Methods("POST")
	router.HandleFunc("/api/nodes", s.Handler.HandleGetNodes).Methods("GET")

	// GitHub Webhook endpoint
	router.HandleFunc("/webhook/github", s.Handler.HandleGitHubWebhook).Methods("POST")

	// Serve static files for the web interface
	// This will serve files from the 'static' directory within 'internal/web'
	staticFileServer := http.FileServer(http.Dir("./internal/web/static"))
	router.PathPrefix("/").Handler(staticFileServer)

	addr := ":" + s.Port
	log.Printf("Web server starting on %s (HTTP enabled)", addr)

	// Revert to simple HTTP server
	srv := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Use ListenAndServe for HTTP
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Could not start server: %v", err)
	}
}
