// Package main is the entry point of the application
package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/Ajnasz/sekret.link/api"
	"github.com/Ajnasz/sekret.link/internal/config"
	"github.com/Ajnasz/sekret.link/internal/durable"
	"github.com/Ajnasz/sekret.link/internal/key"
	"github.com/Ajnasz/sekret.link/internal/models"
	"github.com/Ajnasz/sekret.link/internal/models/migrate"
	"github.com/Ajnasz/sekret.link/internal/services"
)

var (
	version string
	build   string
)

func shutDown(shutdowns ...func() error) chan error {
	errChan := make(chan error)

	var wg sync.WaitGroup
	for i, shutdown := range shutdowns {
		wg.Add(1)
		go func(_ int, shutdown func() error) {
			defer wg.Done()
			if err := shutdown(); err != nil {
				errChan <- err
			}
		}(i, shutdown)
	}

	go func() {
		wg.Wait()
		close(errChan)
	}()

	return errChan
}

func scheduleDeleteExpired(ctx context.Context, db *sql.DB) error {
	manager := services.NewExpiredEntryManager(db, &models.EntryModel{}, &models.EntryKeyModel{})
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			slog.Info("Stop deleting expired entries")
			return nil
		case <-ticker.C:
			if err := manager.DeleteExpired(ctx); err != nil {
				fmt.Fprintf(os.Stderr, "error: %s", err)
			}
		}
	}
}

func removeSocketFile(unixSocket string) {
	if unixSocket == "" {
		return
	}

	stat, err := os.Stat(unixSocket)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// Nothing to clean up: first run, or the previous shutdown
			// already unlinked the socket.
			return
		}
		slog.Error("Failed to stat unix socket", "error", err)
		os.Exit(1)
	}

	if stat.Mode()&os.ModeSocket == 0 {
		slog.Error("Unix socket path exists but is not a socket", "path", unixSocket)
		os.Exit(1)
	}

	// Remove a socket file left behind by a previous run/crash; net.Listen
	// fails with "address already in use" otherwise.
	if err := os.Remove(unixSocket); err != nil {
		slog.Error("Failed to remove unix socket", "error", err)
		os.Exit(1)
	}
}

func listen(ctx context.Context, handlerConfig api.HandlerConfig, unixSocket string) *http.Server {
	mux := http.NewServeMux()

	apiRoot := getAPIRoot(handlerConfig.WebExternalURL)

	secretHandler := api.NewSecretHandler(handlerConfig)
	secretHandler.RegisterHandlers(mux, apiRoot)

	httpServer := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
	}
	network, address := "tcp", httpServer.Addr
	if unixSocket != "" {
		network, address = "unix", unixSocket
	}

	go func() {
		slog.Info("Start listening", "network", network, "address", address, "path", apiRoot)

		if network == "unix" {
			removeSocketFile(address)
		}

		ln, err := net.Listen(network, address)
		if err != nil {
			slog.Error("Failed to listen", "error", err)
			os.Exit(1)
		}

		if network == "unix" {
			// A reverse proxy (e.g. nginx, running as a different user) needs to
			// connect to this socket; net.Listen creates it with permissions
			// derived from umask, which usually excludes other users.
			if err := os.Chmod(address, 0660); err != nil {
				slog.Error("Failed to chmod unix socket", "error", err)
				os.Exit(1)
			}
		}

		if err := httpServer.Serve(ln); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				slog.Error("Failed to serve", "error", err)
				os.Exit(1)
			}
		}
	}()

	return httpServer
}

func getAPIRoot(webExternalURL *url.URL) string {
	apiRoot := ""

	if webExternalURL.Path != "" {
		apiRoot = webExternalURL.Path
	}

	if apiRoot == "" {
		return "/"
	}

	apiRoot = path.Clean(apiRoot)

	if !strings.HasSuffix(apiRoot, "/") {
		apiRoot += "/"
	}

	return apiRoot
}

func getCORSOrigins(list string) []string {
	if list == "" {
		return []string{}
	}

	listSlice := strings.Split(list, ",")
	out := make([]string, 0, len(listSlice))

	for _, origin := range listSlice {
		url, err := url.Parse(origin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: invalid CORS origin: %s", origin)
			continue
		}
		out = append(out, url.String())
	}

	return out
}

func getConfig(ctx context.Context) (*api.HandlerConfig, string, error) {
	var (
		externalURLParam string
		corsOriginsParam string
		expireSeconds    int
		maxExpireSeconds int
		postgresDB       string
		maxDataSize      int64
		queryVersion     bool
		base62Encoding   bool
		unixSocket       string
	)
	flag.StringVar(&externalURLParam, "webExternalURL", "", "Web server external url")
	flag.StringVar(&corsOriginsParam, "corsOrigins", "", "Comma separated list of allowed CORS origins")
	flag.StringVar(&postgresDB, "postgresDB", "", "Connection string for postgresql database backend")
	flag.IntVar(&expireSeconds, "expireSeconds", 60*60*24*7, "Default expiration time in seconds")
	flag.IntVar(&maxExpireSeconds, "maxExpireSeconds", 60*60*24*30, "Max expiration time in seconds")
	flag.Int64Var(&maxDataSize, "maxDataSize", 1024*1024, "Max data size")
	flag.BoolVar(&queryVersion, "version", false, "Get version information")
	flag.BoolVar(&base62Encoding, "base62", false, "Use base62 encoding")
	flag.StringVar(&unixSocket, "unixSocket", "", "Listen on this Unix domain socket path instead of TCP :8080 (for use behind a reverse proxy on the same host)")
	flag.Parse()

	if queryVersion {
		fmt.Printf("%s %s", version, build)
		fmt.Println()
		os.Exit(0)
	}

	extURL, err := url.Parse(externalURLParam)

	if err != nil {
		return nil, "", err
	}

	if base62Encoding {
		key.SetEncodingType(key.Base62Encoding)
	}

	handlerConfig := api.HandlerConfig{
		ExpireSeconds:    expireSeconds,
		MaxExpireSeconds: maxExpireSeconds,
		MaxDataSize:      maxDataSize,
	}

	if maxExpireSeconds < expireSeconds {
		return nil, "", fmt.Errorf("`expireSeconds` must be less or equal then `maxExpireSeconds`")
	}
	handlerConfig.WebExternalURL = extURL

	db, err := durable.OpenDatabaseClient(ctx, config.GetConnectionString(postgresDB))

	if err != nil {
		return nil, "", err
	}
	if err := migrate.PrepareDatabase(ctx, db); err != nil {
		return nil, "", err
	}
	handlerConfig.DB = db
	handlerConfig.CORSOrigins = getCORSOrigins(corsOriginsParam)

	return &handlerConfig, unixSocket, nil
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	// Base context for in-flight requests. Deliberately not the signal
	// context: requests must survive SIGTERM through Shutdown's drain window
	// (a one-time read canceled mid-flight burns the secret), so this is only
	// canceled once the drain is over.
	serverCtx, serverCancel := context.WithCancel(context.Background())
	defer serverCancel()

	handlerConfig, unixSocket, err := getConfig(ctx)

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s", err)
		os.Exit(1)
	}
	go func() {
		err := scheduleDeleteExpired(ctx, handlerConfig.DB)
		if err != nil {
			slog.Error("Error deleting expired entries", "error", err)
		}
	}()
	httpServer := listen(serverCtx, *handlerConfig, unixSocket)

	<-ctx.Done()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	shutdownErrors := shutDown(func() error {
		ctx, cancel := context.WithTimeout(ctx, time.Second*5)
		defer cancel()
		err := httpServer.Shutdown(ctx)
		// Drain finished or timed out: abort any request still running so it
		// can't outlive the server.
		serverCancel()
		return err
	}, func() error {
		return handlerConfig.DB.Close()
	})

	errored := false
	for {
		select {
		case err, ok := <-shutdownErrors:
			if !ok {
				shutdownErrors = nil
			} else if err != nil {
				errored = true
				fmt.Fprintf(os.Stderr, "error: %s", err)
			}
		case <-time.After(time.Second * 15):
			fmt.Fprint(os.Stderr, "error: force quit")
			os.Exit(2)
		}

		if shutdownErrors == nil {
			if errored {
				os.Exit(1)
			}
			return
		}
	}
}
