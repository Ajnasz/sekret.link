package main

import (
	"context"
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
	"github.com/Ajnasz/sekret.link/api/middlewares"
	"github.com/Ajnasz/sekret.link/config"
	"github.com/Ajnasz/sekret.link/storage"
	"github.com/Ajnasz/sekret.link/storage/postgresql"
)

var (
	version string
)

func getStorage(postgresDB string) *postgresql.Storage {
	return postgresql.NewStorage(config.GetConnectionString(postgresDB))
}

func shutDown(shutdowns ...func() error) chan error {
	errChan := make(chan error)

	var wg sync.WaitGroup
	for i, shutdown := range shutdowns {
		wg.Add(1)
		go func(i int, shutdown func() error) {
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

func scheduleDeleteExpired(ctx context.Context, entryStorage storage.Writer) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Second):
			entryStorage.DeleteExpired(ctx)
		}
	}
}

func listen(handlerConfig api.HandlerConfig) *http.Server {
	apiRoot := getAPIRoot(handlerConfig.WebExternalURL)
	fmt.Println("Handle Path: ", apiRoot)

	r := http.NewServeMux()
	r.Handle(
		apiRoot,
		http.StripPrefix(
			apiRoot,
			middlewares.SetupLogging(
				middlewares.SetupHeaders(api.NewSecretHandler(handlerConfig)),
			),
		),
	)
	httpServer := &http.Server{
		Addr:         ":8080",
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
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

	config := api.HandlerConfig{
		ExpireSeconds:    expireSeconds,
		MaxExpireSeconds: maxExpireSeconds,
		MaxDataSize:      maxDataSize,
	}

	if maxExpireSeconds < expireSeconds {
		return nil, fmt.Errorf("`expireSeconds` must be less or equal then `maxExpireSeconds`")
	}
	config.WebExternalURL = extURL

	entryStorage := getStorage(postgresDB)
	if entryStorage == nil {
		return nil, fmt.Errorf("no database backend selected")
	}

	config.EntryStorage = entryStorage
	config.DB = entryStorage.GetDB()

	return &config, nil
}

func main() {
	handlerConfig, err := getConfig()

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	go scheduleDeleteExpired(ctx, handlerConfig.EntryStorage)
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
		return handlerConfig.EntryStorage.Close()
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
