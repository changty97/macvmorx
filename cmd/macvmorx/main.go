package main

import (
	"io"
	"log"
	"os"

	"github.com/changty97/macvmorx/internal/api"
	"github.com/changty97/macvmorx/internal/config"
	"github.com/changty97/macvmorx/internal/core"
	"github.com/changty97/macvmorx/internal/heartbeat"
	"github.com/changty97/macvmorx/internal/store"
	"github.com/changty97/macvmorx/internal/web"
	"github.com/spf13/cobra"
)

var cfg *config.Config // Global config variable

func init() {
	// Load configuration early
	cfg = config.LoadConfig()

	// Use config values as defaults for Cobra flags
	rootCmd.PersistentFlags().StringVarP(&cfg.WebPort, "port", "p", cfg.WebPort, "Port for the web server")
	rootCmd.PersistentFlags().DurationVar(&cfg.OfflineTimeout, "offline-timeout", cfg.OfflineTimeout, "Duration after which a node is considered offline if no heartbeat is received")
	rootCmd.PersistentFlags().DurationVar(&cfg.MonitorInterval, "monitor-interval", cfg.MonitorInterval, "Interval for checking offline nodes")
	rootCmd.PersistentFlags().StringVar(&cfg.GitHubWebhookSecret, "github-webhook-secret", cfg.GitHubWebhookSecret, "GitHub Webhook secret for validation")
	rootCmd.PersistentFlags().StringVar(&cfg.GitHubRunnerRegistrationToken, "github-runner-registration-token", cfg.GitHubRunnerRegistrationToken, "Static GitHub Actions runner registration token")
	rootCmd.PersistentFlags().StringVar(&cfg.LogFilePath, "log-file", cfg.LogFilePath, "Path to a file to write logs (e.g., /var/log/macvmorx.log)") // NEW: Log file flag
	// Removed mTLS flags
}

var rootCmd = &cobra.Command{
	Use:   "macvmorx",
	Short: "macvmorx is a Kubernetes-like orchestrator for Mac virtual machines.",
	Long: `A comprehensive orchestrator for managing Mac virtual machines on Mac Mini labs.
It handles heartbeats, monitors node health, and provides a web interface for easy access.
Now integrates with GitHub webhooks for reactive VM provisioning.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Default command action: start the web server
		startServer()
	},
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the macvmorx web server",
	Run: func(cmd *cobra.Command, args []string) {
		startServer()
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Get the current status of all nodes",
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("To get node status, please access the web interface.")
		log.Printf("Web server will be available at http://localhost:%s", cfg.WebPort) // Changed to HTTP
	},
}

func startServer() {
	// NEW: Configure logging to file if LogFilePath is provided
	if cfg.LogFilePath != "" {
		logFile, err := os.OpenFile(cfg.LogFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatalf("Failed to open log file %s: %v", cfg.LogFilePath, err)
		}
		defer logFile.Close()
		// Write logs to both file and console
		mw := io.MultiWriter(os.Stdout, logFile)
		log.SetOutput(mw)
		log.Printf("Logging to file: %s", cfg.LogFilePath)
	}

	// Initialize the node store
	nodeStore := store.NewNodeStore(cfg.OfflineTimeout)

	// Initialize the job store (NEW)
	jobStore := store.NewJobStore()

	// Initialize the heartbeat processor (pass jobStore now)
	hp := heartbeat.NewProcessor(nodeStore, jobStore)
	go hp.StartOfflineMonitor(cfg.MonitorInterval)

	// Initialize the core orchestrator (pass jobStore now)
	orch, err := core.NewOrchestrator(nodeStore, jobStore, cfg)
	if err != nil {
		log.Fatalf("Failed to initialize orchestrator: %v", err)
	}

	// Initialize API handlers with all dependencies (pass jobStore now)
	apiHandlers := api.NewHandlers(cfg, hp, nodeStore, jobStore, orch)

	// Initialize and start the web server (no mTLS config needed now)
	ws := web.NewServer(cfg.WebPort, apiHandlers, cfg)
	ws.Start()
}

func main() {
	rootCmd.AddCommand(serverCmd)
	rootCmd.AddCommand(statusCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("Error executing command: %v", err)
		os.Exit(1)
	}
}
