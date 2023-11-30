package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"
	"vinr.eu/go-framework/app"
	"vinr.eu/go-framework/database"
	"vinr.eu/go-framework/handler"
	"vinr.eu/go-framework/log"
	"vinr.eu/go-framework/server"
	"vinr.eu/go-framework/user-service/internal/code"
	"vinr.eu/go-framework/user-service/internal/usecase/managing"
	managingtypes "vinr.eu/go-framework/user-service/pkg/api/managing"
)

func main() {
	repository, err := database.NewMongoDBRepository(10*time.Second, "userService")
	if err != nil {
		os.Exit(1)
	}
	defer repository.Disconnect()

	idleConnectionsClosed := make(chan struct{})

	mux := http.NewServeMux()

	mux.Handle("/managing/queries/view-user/",
		handler.NewQueryByIDHandler[managingtypes.ViewUserResponse](repository, managing.ViewUser, handleAppError, mapHeaders))

	server.StartHttpServer(mux, idleConnectionsClosed)

	<-idleConnectionsClosed
}

func handleAppError(appErr app.Error, w http.ResponseWriter, r *http.Request) {
	logger := log.NewLogger(slog.String("traceId", r.Header.Get("x-trace-id")))
	logger.Error("Execution failed", "code", appErr.GetCode(), "err", appErr.Error(),
		"stackTrace", appErr.GetStackTrace(), log.AttrKeyTeam, log.AttrTeamOps)

	// Set API response based on app error.
	switch appErr.GetCode() {
	case code.ErrCode101DataFetchFailed:
		w.WriteHeader(http.StatusNotFound)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func mapHeaders(r *http.Request) map[string]string {
	return map[string]string{"traceId": r.Header.Get("x-trace-id")}
}
