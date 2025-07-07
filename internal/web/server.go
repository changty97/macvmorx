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

// corsMiddleware adds CORS headers to allow cross-origin requests.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow requests from any origin during development.
		// In production, you might want to restrict this to specific origins.
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Requested-With, Authorization")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Start runs the HTTP server.
func (s *Server) Start() {
	router := mux.NewRouter()

	// Apply CORS middleware to all API routes
	apiRouter := router.PathPrefix("/api").Subrouter()
	apiRouter.Use(corsMiddleware) // Apply CORS middleware here

	// API routes
	apiRouter.HandleFunc("/heartbeat", s.Handler.HandleHeartbeat).Methods("POST")
	apiRouter.HandleFunc("/nodes", s.Handler.HandleGetNodes).Methods("GET")
	apiRouter.HandleFunc("/jobs", s.Handler.HandleGetJobs).Methods("GET") // NEW: Jobs API endpoint

	// GitHub Webhook endpoint (can also be behind CORS if needed, but typically not for webhooks)
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
		Handler:      router, // Use the router with middleware
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Use ListenAndServe for HTTP
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Could not start server: %v", err)
	}
}
