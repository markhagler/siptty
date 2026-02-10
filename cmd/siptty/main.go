package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/siptty/siptty/internal/config"
	"github.com/siptty/siptty/internal/engine"
	"github.com/siptty/siptty/internal/tui"
)

func main() {
	configPath := flag.String("config", "", "path to config file")
	flag.Parse()

	// Find config file.
	cfgPath := *configPath
	if cfgPath == "" {
		var err error
		cfgPath, err = config.FindConfigFile()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	}

	// Load config.
	cfg, err := config.Load(cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	// Set up structured logging.
	logLevel := slog.LevelInfo
	switch {
	case cfg.General.LogLevel <= 1:
		logLevel = slog.LevelError
	case cfg.General.LogLevel <= 2:
		logLevel = slog.LevelWarn
	case cfg.General.LogLevel <= 3:
		logLevel = slog.LevelInfo
	default:
		logLevel = slog.LevelDebug
	}

	logOpts := &slog.HandlerOptions{Level: logLevel}
	var logHandler slog.Handler
	if cfg.General.LogFile != "" {
		f, err := os.OpenFile(cfg.General.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error opening log file: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()
		logHandler = slog.NewTextHandler(f, logOpts)
	} else {
		logHandler = slog.NewTextHandler(os.Stderr, logOpts)
	}
	slog.SetDefault(slog.New(logHandler))

	// Create engine.
	eng, err := engine.NewEngine(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	// Create TUI.
	app := tui.NewApp(eng)

	// Start engine.
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := eng.Start(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	// Run TUI (blocks until quit).
	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "TUI error: %v\n", err)
	}

	// Clean shutdown.
	eng.Stop()
}
