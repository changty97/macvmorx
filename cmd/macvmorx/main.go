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
}

var rootCmd = &cobra.Command{
	Use:   "macvmorx",
	Short: "macvmorx is a Kubernetes-like orchestrator for Mac virtual machines.",
	Long: `A comprehensive orchestrator for managing Mac virtual machines on Mac Mini labs.
It handles heartbeats, monitors node health, and provides a web interface for easy access.`,
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
		// This would typically query the running macvmorx server's API
		// For simplicity, we'll just print a message for now.
		// In a real CLI, you'd make an HTTP request to http://localhost:webPort/api/nodes
		log.Println("To get node status, please access the web interface or implement API client logic here.")
		log.Printf("Web server will be available at http://localhost:%s", cfg.WebPort)
	},
}

func startServer() {
	// Initialize the node store with a configurable offline timeout from config
	nodeStore := store.NewNodeStore(cfg.OfflineTimeout)

	// Initialize the heartbeat processor
	hp := heartbeat.NewProcessor(nodeStore)

	// Start the offline monitor in a goroutine
	go hp.StartOfflineMonitor(cfg.MonitorInterval)

	// Initialize the core orchestrator
	orch := core.NewOrchestrator(nodeStore)

	// Initialize API handlers with the new orchestrator dependency
	apiHandlers := api.NewHandlers(hp, nodeStore, orch)

	// Initialize and start the web server using config port
	ws := web.NewServer(cfg.WebPort, apiHandlers)
	ws.Start()
}

func main() {
	rootCmd.AddCommand(serverCmd)
	rootCmd.AddCommand(statusCmd) // Example CLI command

	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("Error executing command: %v", err)
		os.Exit(1)
	}
}
