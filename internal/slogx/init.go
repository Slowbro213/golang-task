package slogx

import (
	"echo-app/internal/config"
	"fmt"
	"io"
	"log/slog"
	"os"
)

func Init(conf config.Log) (err error) {
	writer := io.Writer(os.Stdout)
	if conf.File != "" {
		const permission = 0o644

		writer, err = os.OpenFile(conf.File, os.O_APPEND|os.O_CREATE|os.O_WRONLY, permission)
		if err != nil {
			return fmt.Errorf("open file %s: %w", conf.File, err)
		}
	}

	level := slog.LevelDebug
	if conf.Level != "" {
		if err = level.UnmarshalText([]byte(conf.Level)); err != nil {
			return fmt.Errorf("parse log level %s: %w", conf.Level, err)
		}
	}

	hostname, err := os.Hostname()
	if err != nil {
		return fmt.Errorf("get hostname: %w", err)
	}

	jsonHandler := slog.NewJSONHandler(writer, &slog.HandlerOptions{AddSource: conf.AddSource, Level: level})

	traceHandler := newTraceHandler(jsonHandler)

	logger := slog.New(traceHandler).
		With("application", conf.Application).
		With("hostname", hostname)

	slog.SetDefault(logger)

	return nil
}
