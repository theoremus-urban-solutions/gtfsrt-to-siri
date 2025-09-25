package gtfsrtsiri

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	server *http.Server
)

func StartServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", handleHealth)
	mux.HandleFunc("/api/siri/vehicle-monitoring.json", handleVehicleMonitoringJSON)
	mux.HandleFunc("/api/siri/stop-monitoring.json", handleStopMonitoringJSON)
	mux.HandleFunc("/api/siri/vehicle-monitoring.xml", handleVehicleMonitoringXML)
	mux.HandleFunc("/api/siri/stop-monitoring.xml", handleStopMonitoringXML)

	addr := fmt.Sprintf(":%d", Config.Server.Port)
	server = &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()
	log.Printf("server listening on %s", addr)
}

func HandleGracefulShutdown() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
	log.Printf("shutdown signal received")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if server != nil {
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("server shutdown error: %v", err)
		} else {
			log.Printf("server shut down successfully")
		}
	}
}
