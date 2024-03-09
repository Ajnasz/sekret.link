// Package main is the entry point of the application
package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
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
	"github.com/Ajnasz/sekret.link/config"
	"github.com/Ajnasz/sekret.link/internal/durable"
	"github.com/Ajnasz/sekret.link/internal/models"
	"github.com/Ajnasz/sekret.link/internal/services"
)

var (
	version string
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
	manager := services.NewExpiredEntryManager(db, &models.EntryModel{})
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(time.Second):
			manager.DeleteExpired(ctx)
		}
	}
}

func listen(handlerConfig api.HandlerConfig) *http.Server {
	mux := http.NewServeMux()

	apiRoot := getAPIRoot(handlerConfig.WebExternalURL)

	secretHandler := api.NewSecretHandler(handlerConfig)
	secretHandler.RegisterHandlers(mux, apiRoot)

	httpServer := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		fmt.Println("Handle Path: ", apiRoot)
		if err := httpServer.ListenAndServe(); err != nil {
			if err.Error() != "http: Server closed" {
				fmt.Fprintf(os.Stderr, "error: %s", err)
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

func getConfig() (*api.HandlerConfig, error) {
	var (
		externalURLParam string
		expireSeconds    int
		maxExpireSeconds int
		postgresDB       string
		maxDataSize      int64
		queryVersion     bool
	)
	flag.StringVar(&externalURLParam, "webExternalURL", "", "Web server external url")
	flag.StringVar(&postgresDB, "postgresDB", "", "Connection string for postgresql database backend")
	flag.IntVar(&expireSeconds, "expireSeconds", 60*60*24*7, "Default expiration time in seconds")
	flag.IntVar(&maxExpireSeconds, "maxExpireSeconds", 60*60*24*30, "Max expiration time in seconds")
	flag.Int64Var(&maxDataSize, "maxDataSize", 1024*1024, "Max data size")
	flag.BoolVar(&queryVersion, "version", false, "Get version information")
	flag.Parse()

	if queryVersion {
		fmt.Println(version)
		os.Exit(0)
	}

	extURL, err := url.Parse(externalURLParam)

	if err != nil {
		return nil, err
	}

	handlerConfig := api.HandlerConfig{
		ExpireSeconds:    expireSeconds,
		MaxExpireSeconds: maxExpireSeconds,
		MaxDataSize:      maxDataSize,
	}

	if maxExpireSeconds < expireSeconds {
		return nil, fmt.Errorf("`expireSeconds` must be less or equal then `maxExpireSeconds`")
	}
	handlerConfig.WebExternalURL = extURL

	db, err := durable.OpenDatabaseClient(context.Background(), config.GetConnectionString(postgresDB))

	if err != nil {
		return nil, err
	}
	handlerConfig.DB = db

	return &handlerConfig, nil
}

func main() {
	handlerConfig, err := getConfig()

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	go scheduleDeleteExpired(ctx, handlerConfig.DB)
	httpServer := listen(*handlerConfig)

	termChan := make(chan os.Signal)
	signal.Notify(termChan, syscall.SIGTERM, syscall.SIGINT)

	defer close(termChan)
	<-termChan
	cancel()

	shutdownErrors := shutDown(func() error {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		return httpServer.Shutdown(ctx)
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
