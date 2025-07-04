package web

import (
	"log"
	"net/http"
	"time"

	"github.com/changty97/macvmorx/internal/api"
	"github.com/gorilla/mux" // Using gorilla/mux for more advanced routing
)

// Server represents the web server for macvmorx.
type Server struct {
	Port    string
	Handler *api.Handlers
}

// NewServer creates a new web server instance.
func NewServer(port string, handler *api.Handlers) *Server {
	return &Server{
		Port:    port,
		Handler: handler,
	}
}

// Start runs the HTTP server.
func (s *Server) Start() {
	router := mux.NewRouter()

	// API routes
	router.HandleFunc("/api/heartbeat", s.Handler.HandleHeartbeat).Methods("POST")
	router.HandleFunc("/api/nodes", s.Handler.HandleGetNodes).Methods("GET")
	router.HandleFunc("/api/schedule-vm", s.Handler.HandleScheduleVM).Methods("POST") // New API endpoint

	// Serve static files for the web interface
	// This will serve files from the 'static' directory within 'internal/web'
	staticFileServer := http.FileServer(http.Dir("./internal/web/static"))
	router.PathPrefix("/").Handler(staticFileServer)

	addr := ":" + s.Port
	log.Printf("Web server starting on %s", addr)

	srv := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Could not start server: %v", err)
	}
}
