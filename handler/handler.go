package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"vinr.eu/go-framework/app"
	"vinr.eu/go-framework/database"
	"vinr.eu/go-framework/log"
)

type CommandHandlerFunc[T any] func(*database.Repository, T, map[string]string) *app.Error
type CommandWithResponseHandlerFunc[T any, R any] func(*database.Repository, T, map[string]string) (*R, *app.Error)
type QueryHandlerFunc[T any] func(*database.Repository, map[string]string) (*T, *app.Error)
type QueryByIDHandlerFunc[T any] func(*database.Repository, string, map[string]string) (*T, *app.Error)
type QueryByParamsHandlerFunc[T any] func(*database.Repository, map[string]string, map[string]string) (*T, *app.Error)
type AppErrorHandlerFunc func(appErr app.Error, w http.ResponseWriter, r *http.Request)
type AuthHeaderMapFunc func(r *http.Request) map[string]string

func NewCommandHandler[T any](repo *database.Repository, handlerFunc CommandHandlerFunc[T],
	appErrorhandlerFunc AppErrorHandlerFunc, authHeaderMapFunc AuthHeaderMapFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var command T
		if err := json.NewDecoder(r.Body).Decode(&command); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		appErr := handlerFunc(repo, command, authHeaderMapFunc(r))
		if appErr != nil {
			appErrorhandlerFunc(*appErr, w, r)
			return
		}
	})
}

func NewCommandWithResponseHandler[T any, R any](repo *database.Repository, handlerFunc CommandWithResponseHandlerFunc[T, R],
	appErrorhandlerFunc AppErrorHandlerFunc, authHeaderMapFunc AuthHeaderMapFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := log.NewLogger()
		var command T
		if err := json.NewDecoder(r.Body).Decode(&command); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		response, appErr := handlerFunc(repo, command, authHeaderMapFunc(r))
		if appErr != nil {
			appErrorhandlerFunc(*appErr, w, r)
			return
		}
		if response != nil {
			setResponse(w, response, logger)
		}
	})
}

func NewQueryByParamsHandler[T any](repo *database.Repository, handlerFunc QueryByParamsHandlerFunc[T],
	appErrorhandlerFunc AppErrorHandlerFunc, authHeaderMapFunc AuthHeaderMapFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := log.NewLogger()
		response, appErr := handlerFunc(repo, extractQueryParams(r), authHeaderMapFunc(r))
		if appErr != nil {
			appErrorhandlerFunc(*appErr, w, r)
			return
		}
		setResponse(w, response, logger)
	})
}

func NewQueryByIDHandler[T any](repo *database.Repository, handlerFunc QueryByIDHandlerFunc[T],
	appErrorhandlerFunc AppErrorHandlerFunc, authHeaderMapFunc AuthHeaderMapFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := log.NewLogger()
		response, appErr := handlerFunc(repo, getIdFromPath(r), authHeaderMapFunc(r))
		if appErr != nil {
			appErrorhandlerFunc(*appErr, w, r)
			return
		}
		setResponse(w, response, logger)
	})
}

func NewQueryHandler[T any](repo *database.Repository, handlerFunc QueryHandlerFunc[T],
	appErrorhandlerFunc AppErrorHandlerFunc, authHeaderMapFunc AuthHeaderMapFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := log.NewLogger()
		response, appErr := handlerFunc(repo, authHeaderMapFunc(r))
		if appErr != nil {
			appErrorhandlerFunc(*appErr, w, r)
			return
		}
		setResponse(w, response, logger)
	})
}

func setResponse[R any](w http.ResponseWriter, response *R, logger *slog.Logger) {
	jsonData, err := json.Marshal(response)
	if err != nil {
		// No need for error codes here. If this fails then the API itself fails and needs backend investigation.
		logger.Error("Response marshal failed", "err", err, log.AttrKeyTeam, log.AttrTeamDev)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(jsonData)
	if err != nil {
		logResponseWriteFailure(logger, err)
	}
}

func logResponseWriteFailure(logger *slog.Logger, err error) {
	var netErr *net.OpError
	if errors.As(err, &netErr) && netErr.Err.Error() == "write: broken pipe" {
		logger.Warn("Response write failed due to broken pipe")
	} else {
		logger.Error("Response write failed", "err", err, log.AttrKeyTeam, log.AttrTeamDev)
	}
}

func getIdFromPath(r *http.Request) string {
	// Extract the Query from the URL path parameter
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		return ""
	}
	return parts[4]
}

func extractQueryParams(r *http.Request) map[string]string {
	params := make(map[string]string)
	for key, values := range r.URL.Query() {
		// If multiple values for the same key exist, take the first value
		params[key] = values[0]
	}
	return params
}
