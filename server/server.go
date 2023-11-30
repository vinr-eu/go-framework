package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"
	"vinr.eu/go-framework/log"
)

type CleanupFunc func()

func StartHttpServer(mux *http.ServeMux, idleConnectionsClosed chan struct{}) {
	logger := log.NewLogger()

	// Create http server
	address := ""
	serverPort := os.Getenv("SERVER_PORT")
	if serverPort == "" {
		serverPort = "8080"
	}
	serverAddr := fmt.Sprintf("%s:%s", address, serverPort)
	srv := &http.Server{Addr: serverAddr, Handler: mux}

	// Prepare http server for graceful shutdown.
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint

		// We received an interrupt signal, shut down.
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			// Error from closing listeners, or context timeout.
			logger.Error("Server shutdown failed", "err", err, log.AttrKeyTeam, log.AttrTeamOps)
		}
		logger.Info("Server shutdown")
		close(idleConnectionsClosed)
	}()

	// Start http server.
	go func() {
		logger.Info("Server started")
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			// Error starting or closing listener.
			logger.Error("Server startup failed", "err", err, log.AttrKeyTeam, log.AttrTeamOps)
		}
	}()
}
