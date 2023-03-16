package slug_test

import (
	"context"
	"errors"
	"net/netip"
	"regexp"
	"strings"
	"testing"
	"time"

	log "github.com/dotse/slug"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/slog"
)

func TestHandler(t *testing.T) {
	t.Parallel()

	var (
		ctx = context.Background()
		sgr = regexp.MustCompile(`\033\[[\d;]*m`)
	)

	for _, c := range [...]struct {
		Group    string
		Level    slog.Level
		Msg      string
		Attrs    []slog.Attr
		Patterns []string
	}{
		{
			Level: slog.LevelError,
			Msg:   "test message with attributes",
			Attrs: []slog.Attr{
				slog.Any("any", netip.AddrFrom4([4]byte{127, 0, 0, 1})),
				slog.Bool("bool", true),
				slog.Duration("duration", time.Hour+30*time.Minute),
				slog.Any("error", errors.New("test error")),
				slog.Float64("float64", 123_456.789),
				slog.Int("int", -123_456),
				slog.Int64("int64", 42),
				slog.String("string", "a string"),
				slog.Time("time", time.Date(2000, 1, 2, 3, 4, 5, 6, time.UTC)),
				slog.Uint64("uint64", 123456),
			},
			Patterns: []string{
				`^[+,:TZ\d-]{20,} +ERROR +test message with attributes`,
				`\sany=127.0.0.1\b`,
				`\sbool=true\b`,
				`\sduration=1h30m0s\b`,
				`\serror=test error\b`,
				`\sfile=.*/handler_test.go\b`,
				`\sfloat64=123,456.789\b`,
				`\sint64=42\b`,
				`\sint=-123,456\b`,
				`\sline=\d+\b\b`,
				`\sstring=a string\b`,
				`\stime=2000-01-02T03:04:05,000Z\b`,
				`\suint64=123,456\b`,
			},
		},

		{
			Level: slog.LevelWarn,
			Msg:   "group test",
			Attrs: []slog.Attr{
				slog.String("outside", "foo"),
				slog.Group("group",
					slog.String("inside", "bar"),
					slog.Int("inside-int", 42),
				),
			},
			Patterns: []string{
				`^[+,:TZ\d-]{20,} +WARN +group test `,
				`\soutside=foo\b`,
				`\sgroup\.inside=bar\b`,
				`\sgroup\.inside-int=42\b`,
			},
		},

		{
			Group: "group",
			Level: slog.LevelInfo,
			Msg:   "logger with group",
			Attrs: []slog.Attr{
				slog.String("foo", "bar"),
				slog.Group("bar",
					slog.String("foo", "qux"),
				),
			},
			Patterns: []string{
				`^[+,:TZ\d-]{20,} +INFO +logger with group `,
				`\sgroup\.foo=bar\b`,
				`\sgroup\.bar\.foo=qux\b`,
			},
		},

		{
			Level: slog.LevelWarn,
			Msg:   "message\twith\r\nnon-printable\u2003characters",
			Attrs: []slog.Attr{
				slog.String("emoji", "😻"),
			},
			Patterns: []string{
				`^[+,:TZ\d-]{20,} +WARN +message\\twith\\r\\nnon-printableU\+2003characters`,
				`\semoji=😻`,
			},
		},
	} {
		c := c

		t.Run(c.Msg, func(t *testing.T) {
			t.Parallel()

			var (
				buffer strings.Builder
				h      = log.NewHandler(log.HandlerOptions{
					HandlerOptions: slog.HandlerOptions{
						AddSource: true,
					},
				}, &buffer)
			)

			require.NotNil(t, h)

			l := slog.New(h)

			if c.Group != "" {
				l = l.WithGroup(c.Group)
			}

			l.LogAttrs(ctx, c.Level, c.Msg, c.Attrs...)

			str := sgr.ReplaceAllString(buffer.String(), "")

			for _, pattern := range c.Patterns {
				assert.Regexp(t, pattern, str)
			}
		})
	}
}
