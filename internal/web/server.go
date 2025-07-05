package web

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/changty97/macvmorx/internal/api"
	"github.com/changty97/macvmorx/internal/config" // Import config for mTLS paths
	"github.com/gorilla/mux"                        // Using gorilla/mux for more advanced routing
)

// Server represents the web server for macvmorx.
type Server struct {
	Port    string
	Handler *api.Handlers
	Cfg     *config.Config // Add config to access mTLS paths
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
	log.Printf("Web server starting on %s (mTLS enabled)", addr)

	// --- mTLS Configuration for the Orchestrator Server ---
	caCert, err := ioutil.ReadFile(s.Cfg.CACertPath)
	if err != nil {
		log.Fatalf("Failed to read CA certificate for server: %v", err)
	}
	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		log.Fatalf("Failed to append CA certificate to pool.")
	}

	serverCert, err := tls.LoadX509KeyPair(s.Cfg.ServerCertPath, s.Cfg.ServerKeyPath)
	if err != nil {
		log.Fatalf("Failed to load server certificate and key: %v", err)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientCAs:    caCertPool,                     // Trust client certificates signed by our CA
		ClientAuth:   tls.RequireAndVerifyClientCert, // Require and verify client certificates (mTLS)
		MinVersion:   tls.VersionTLS12,               // Enforce TLS 1.2 or higher
	}
	tlsConfig.BuildNameToCertificate() // Build map for faster lookups

	srv := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
		TLSConfig:    tlsConfig, // Apply the mTLS configuration
	}

	// Use ListenAndServeTLS for HTTPS
	if err := srv.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Could not start server: %v", err)
	}
}
