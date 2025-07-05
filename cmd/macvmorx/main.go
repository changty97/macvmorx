package main

import (
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
	// mTLS flags
	rootCmd.PersistentFlags().StringVar(&cfg.CACertPath, "ca-cert-path", cfg.CACertPath, "Path to CA certificate (for mTLS)")
	rootCmd.PersistentFlags().StringVar(&cfg.ServerCertPath, "server-cert-path", cfg.ServerCertPath, "Path to server certificate (for orchestrator's listener)")
	rootCmd.PersistentFlags().StringVar(&cfg.ServerKeyPath, "server-key-path", cfg.ServerKeyPath, "Path to server private key (for orchestrator's listener)")
	rootCmd.PersistentFlags().StringVar(&cfg.ClientCertPath, "client-cert-path", cfg.ClientCertPath, "Path to client certificate (for orchestrator talking to agents)")
	rootCmd.PersistentFlags().StringVar(&cfg.ClientKeyPath, "client-key-path", cfg.ClientKeyPath, "Path to client private key (for orchestrator talking to agents)")
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
		log.Printf("Web server will be available at https://localhost:%s (note HTTPS)", cfg.WebPort)
	},
}

func startServer() {
	// Initialize the node store
	nodeStore := store.NewNodeStore(cfg.OfflineTimeout)

	// Initialize the heartbeat processor
	hp := heartbeat.NewProcessor(nodeStore)
	go hp.StartOfflineMonitor(cfg.MonitorInterval)

	// Initialize the core orchestrator (now requires config for mTLS client setup)
	orch, err := core.NewOrchestrator(nodeStore, cfg)
	if err != nil {
		log.Fatalf("Failed to initialize orchestrator: %v", err)
	}

	// Initialize API handlers with all dependencies (no GitHub client now)
	apiHandlers := api.NewHandlers(cfg, hp, nodeStore, orch)

	// Initialize and start the web server (now requires config for mTLS server setup)
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
