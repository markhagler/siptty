package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"

	"github.com/siptty/siptty/internal/config"
	"github.com/siptty/siptty/internal/engine"
	"github.com/siptty/siptty/internal/tui"
)

func main() {
	configPath := flag.String("config", "", "path to config file")
	debug := flag.Bool("debug", false, "enable pprof endpoint on localhost:6060")
	flag.Parse()

	if *debug {
		go func() {
			slog.Info("pprof server starting", "addr", "localhost:6060")
			if err := http.ListenAndServe("localhost:6060", nil); err != nil {
				slog.Error("pprof server failed", "error", err)
			}
		}()
	}

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
	logFile := cfg.General.LogFile
	if logFile == "" {
		// TUI owns stderr — logs must go to a file, never the terminal.
		logFile = "siptty.log"
	}
	f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error opening log file %s: %v\n", logFile, err)
		os.Exit(1)
	}
	defer f.Close()
	logHandler = slog.NewTextHandler(f, logOpts)
	slog.SetDefault(slog.New(logHandler))

	// Create engine.
	eng, err := engine.NewEngine(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	// Create TUI.
	app := tui.NewApp(eng)

	// Start engine with a cancellable context (not tied to signals).
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := eng.Start(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	// Signal handler: first SIGINT/SIGTERM stops gracefully, second force-exits.
	sigCh := make(chan os.Signal, 2)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		slog.Info("signal received, shutting down")
		app.Stop()
		cancel()
		<-sigCh
		slog.Warn("second signal received, force exiting")
		os.Exit(1)
	}()

	// Run TUI (blocks until quit).
	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "TUI error: %v\n", err)
	}

	// Clean shutdown.
	cancel()
	eng.Stop()
}
