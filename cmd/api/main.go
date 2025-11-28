package main

import (
	"bookcabin-test/internal/core/services"
	"bookcabin-test/internal/handlers"
	"bookcabin-test/internal/platform/providers"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	aggregator := services.NewAggregator([]providers.ProviderInterface{
		&providers.GarudaProvider{},
		&providers.LionAirProvider{},
		&providers.BatikAirProvider{},
		&providers.AirAsiaProvider{},
	})

	searchHandler := handlers.NewSearchHandlers(aggregator)
	srv := &http.Server{
		Addr: ":8080",
		// Daftarkan handler menggunakan ServeMux default
		Handler: http.DefaultServeMux,
		// Set timeout yang disarankan
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	http.HandleFunc("/search", searchHandler.SearchFlight)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		fmt.Println("Server running on port 8080. Press Ctrl+C to stop.")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Could not listen on %s: %v\n", srv.Addr, err)
		}
	}()

	<-stop

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited properly.")
}
