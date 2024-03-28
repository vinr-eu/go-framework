package main

import (
	"github.com/vinr-eu/go-framework/app"
	"github.com/vinr-eu/go-framework/database"
	"github.com/vinr-eu/go-framework/handler"
	"github.com/vinr-eu/go-framework/log"
	"github.com/vinr-eu/go-framework/server"
	"github.com/vinr-eu/go-framework/user-service/internal/code"
	"github.com/vinr-eu/go-framework/user-service/internal/usecase/managing"
	managingtypes "github.com/vinr-eu/go-framework/user-service/pkg/api/managing"
	"log/slog"
	"net/http"
	"os"
	"time"
)

func main() {
	// Create all repositories first and defer the Disconnect.
	repository, err := database.NewMongoDBRepository(10*time.Second, "userService")
	if err != nil {
		os.Exit(1)
	}
	defer repository.Disconnect()

	// For graceful handling of server shutdown.
	idleConnectionsClosed := make(chan struct{})

	mux := http.NewServeMux()

	// Pass the handler for each use case.
	mux.Handle("/managing/queries/view-user/",
		handler.NewQueryByIDHandler[managingtypes.ViewUserResponse](repository, managing.ViewUser, handleAppError, mapHeaders))

	// Start HTTP server to serve requests.
	server.StartHTTPServer(mux, idleConnectionsClosed)

	<-idleConnectionsClosed
}

func handleAppError(appErr app.Error, w http.ResponseWriter, r *http.Request) {
	// Create logger for printing the error and stack trace with more logging context for better troubleshooting.
	logger := log.NewLogger(slog.String("traceId", r.Header.Get("x-trace-id")))
	logger.Error("Execution failed", "code", appErr.GetCode(), "err", appErr.Error(),
		"stackTrace", appErr.GetStackTrace(), log.AttrKeyTeam, log.AttrTeamOps)

	// Set API response based on app error and the code. More specific error handling will help more.
	switch appErr.GetCode() {
	case code.ErrCode101DataFetchFailed:
		w.WriteHeader(http.StatusNotFound)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func mapHeaders(r *http.Request) map[string]string {
	// In case your service has auth tokens or other security headers add them to the map. You will get it in use cases.
	return map[string]string{"traceId": r.Header.Get("x-trace-id")}
}
