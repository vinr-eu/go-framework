package log

import (
	"log/slog"
	"os"
)

type AttrTeam string

const (
	AttrKeyTeam string   = "team"
	AttrTeamDev AttrTeam = "dev"
	AttrTeamSec AttrTeam = "sec"
	AttrTeamOps AttrTeam = "ops"
)

func NewLogger(attrs ...slog.Attr) *slog.Logger {
	// Create new application logger
	var handler slog.Handler
	if os.Getenv("ENVIRONMENT") == "dev" {
		handler = slog.NewTextHandler(os.Stdout, nil)
	} else {
		handler = slog.NewJSONHandler(os.Stdout, nil)
	}
	// If attrs are present then create new with attrs
	if len(attrs) > 0 {
		return slog.New(handler.WithAttrs(attrs))
	}
	return slog.New(handler)
}
