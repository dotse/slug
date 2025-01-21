package slug

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"runtime"
	"strings"
	"sync"

	"github.com/dotse/slug/internal"
	"github.com/logrusorgru/aurora/v4"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

const (
	// DefaultTimeFormat is the default format for any [time.Time] logged.
	DefaultTimeFormat = "2006-01-02T15:04:05,999Z07:00"
)

var (
	_ slog.Handler = (*Handler)(nil)
	_ slog.Leveler = (*Handler)(nil)
)

// Handler is a [slog.Handler] that writes human-readable logs.
type Handler struct {
	shared *shared
	attrs  []slog.Attr
	groups []string
}

// NewHandler returns a new [Handler] with [HandlerOptions].
func NewHandler(options HandlerOptions, w io.Writer) *Handler {
	if options.Level == nil {
		options.Level = slog.LevelInfo
	}

	if options.Language == language.Und {
		options.Language = language.BritishEnglish
	}

	if options.TimeFormat == "" {
		options.TimeFormat = DefaultTimeFormat
	}

	return &Handler{
		shared: &shared{
			Writer:  w,
			options: options,
			printer: message.NewPrinter(options.Language),
		},
		attrs:  []slog.Attr{},
		groups: []string{},
	}
}

// Enabled returns true if a message at a [slog.Level] would be logged.
func (h *Handler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.Level()
}

// Handle handles a [slog.Record].
func (h *Handler) Handle(_ context.Context, r slog.Record) error {
	t := aurora.Faint(r.Time.Format(h.shared.options.TimeFormat))

	l := aurora.Bold(r.Level.String())

	switch {
	case r.Level >= slog.LevelError:
		l = l.Red()

	case r.Level >= slog.LevelWarn:
		l = l.Yellow()

	case r.Level <= slog.LevelDebug:
		l = l.Faint()

	default:
		l = l.Blue()
	}

	h.shared.Lock()
	defer h.shared.Unlock()

	fmt.Fprintf(h.shared, "%s %-5s %s", t, l, internal.Escape(r.Message))

	prefix := strings.Join(append(h.groups, ""), ".")

	for _, attr := range h.attrs {
		h.printAttr(prefix, attr)
	}

	r.Attrs(func(attr slog.Attr) bool {
		h.printAttr(prefix, attr)
		return true
	})

	if h.shared.options.AddSource && r.PC != 0 {
		frame, _ := runtime.CallersFrames([]uintptr{r.PC}).Next()
		h.printAttr("", slog.String("file", frame.File))
		h.printAttr("", slog.Int("line", frame.Line))
		h.printAttr("", slog.String("function", frame.Function))
	}

	fmt.Fprintln(h.shared)

	return nil
}

// Level returns the current [slog.Level].
func (h *Handler) Level() slog.Level {
	return h.shared.options.Level.Level()
}

// WithAttrs returns a hew [Handler] with additional attributes.
func (h *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		return h
	}

	c := h.clone()
	c.attrs = append(c.attrs, attrs...)

	return c
}

// WithGroup returns a hew [Handler] with an additional group.
func (h *Handler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}

	c := h.clone()
	c.groups = append(c.groups, name)

	return c
}

func (h *Handler) clone() *Handler {
	return &Handler{
		shared: h.shared,
		attrs:  h.attrs,
		groups: h.groups,
	}
}

func (h *Handler) printAttr(prefix string, attr slog.Attr) {
	value := attr.Value.Resolve()

	var (
		isErr bool
		val   string
	)

	switch value.Kind() {
	case slog.KindFloat64:
		val = h.shared.printer.Sprint(value.Float64())

	case slog.KindGroup:
		prefix += attr.Key + "."
		for _, attr := range value.Group() {
			h.printAttr(prefix, attr)
		}

		return

	case slog.KindInt64:
		val = h.shared.printer.Sprint(value.Int64())

	case slog.KindTime:
		val = value.Time().Format(h.shared.options.TimeFormat)

	case slog.KindUint64:
		val = h.shared.printer.Sprint(value.Uint64())

	case slog.KindAny:
		_, isErr = value.Any().(error)
		fallthrough

	default:
		val = internal.Escape(value.String())
	}

	var (
		k = aurora.Bold(attr.Key)
		v aurora.Value
	)

	if isErr {
		k = k.Red()
		v = aurora.Red(val)
	} else {
		k = k.Magenta()
		v = aurora.Cyan(val)
	}

	fmt.Fprintf(h.shared, " %s%s=%s", aurora.Magenta(prefix), k, v)
}

// HandlerOptions are options for a [Handler].
type HandlerOptions struct {
	// Use all of [slog.HandlerOptions].
	slog.HandlerOptions

	// Language for formatting numbers.
	Language language.Tag

	// TimeFormat for timestamps and [time.Time] attributes. Defaults to
	// [DefaultTimeFormat] if unset.
	TimeFormat string
}

type shared struct {
	io.Writer
	options HandlerOptions
	printer *message.Printer
	sync.Mutex
}
