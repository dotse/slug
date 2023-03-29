package main

/*

This code is used to generate some example logging.

*/

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/dotse/slug"
	"golang.org/x/exp/slog"
)

func main() {
	var (
		ctx = context.Background()
		h   = slug.NewHandler(slug.HandlerOptions{
			HandlerOptions: slog.HandlerOptions{
				Level: slog.LevelDebug,
			},
		}, os.Stdout)
	)

	fmt.Println()

	for i, s := range [...]struct {
		Message string
		Level   slog.Level
		Attrs   []slog.Attr
	}{
		{
			Message: "Starting up",
			Level:   slog.LevelDebug,
		},

		{
			Message: "Reading...",
			Level:   slog.LevelInfo,
			Attrs: []slog.Attr{
				slog.Group("message",
					slog.Int("count", 42),
				),
				slog.String("source", "localhost"),
			},
		},

		{
			Message: "Invalid content",
			Level:   slog.LevelWarn,
			Attrs: []slog.Attr{
				slog.Group("message",
					slog.Int("id", 17),
					slog.String("content", "\uFEFFhello\n"),
				),
			},
		},

		{
			Message: "Parsing failed",
			Level:   slog.LevelError,
			Attrs: []slog.Attr{
				slog.Any("error", errors.New("unsupported sequence")),
			},
		},
	} {
		fmt.Print("  ")

		r := slog.Record{
			Time:    time.Date(2023, time.April, 9, 12, 34, 56+i, 0, time.UTC),
			Message: s.Message,
			Level:   s.Level,
		}

		r.AddAttrs(s.Attrs...)

		h.Handle(ctx, r)
	}

	fmt.Println()
}
