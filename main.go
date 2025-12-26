package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Address  string `yaml:"address"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Timeout  int    `yaml:"timeout_seconds"`
}

// debugBluHandler returns an HTTP handler that writes the parsed blu map as JSON.
func debugBluHandler(config Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bluMap, err := fetchComponentsMap(config)
		if err != nil {
			http.Error(w, fmt.Sprintf("error fetching blu components: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		if err := enc.Encode(bluMap); err != nil {
			http.Error(w, fmt.Sprintf("error encoding JSON: %v", err), http.StatusInternalServerError)
			return
		}
	}
}

func main() {
	// Read configuration
	config, err := readConfig("/etc/shelly-blu-trv-exporter/config.yaml")
	if err != nil {
		log.Fatalf("Error reading configuration: %v", err)
	}

	if config.Timeout == 0 {
		config.Timeout = 5 // Default 5 seconds
	}

	// Validate configuration
	if config.Address == "" || config.Username == "" || config.Password == "" {
		log.Fatal("Missing required configuration fields")
	}

	log.Println("Monitoring Shelly Blu TRV at", config.Address)

	// Register Blu collector
	bluCollector := NewBluCollector(config)
	prometheus.MustRegister(bluCollector)
	// Unregister default collectors we don't need
	prometheus.Unregister(collectors.NewGoCollector())
	prometheus.Unregister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))

	// Start Prometheus HTTP server
	http.Handle("/metrics", promhttp.Handler())
	// Debug endpoint that returns parsed blu map
	http.HandleFunc("/debug/blu", debugBluHandler(config))
	go func() {
		log.Println("Starting Prometheus exporter on :8080/metrics and /debug/blu")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Graceful shutdown handling
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	<-stop
	log.Println("Shutting down...")
}

func readConfig(filename string) (Config, error) {
	var config Config

	data, err := os.ReadFile(filename)
	if err != nil {
		return config, err
	}

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return config, err
	}

	return config, nil
}
