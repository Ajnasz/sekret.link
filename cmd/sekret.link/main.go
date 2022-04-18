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
	"github.com/Ajnasz/sekret.link/config"
	"github.com/Ajnasz/sekret.link/storage"
)

var (
	version string
)

func getStorage(postgresDB string) storage.VerifyStorage {
	return storage.NewStorage(config.GetConnectionString(postgresDB))
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

func scheduleDeleteExpired(entryStorage storage.VerifyStorage, stopChan chan interface{}) {
	for {
		select {
		case <-time.After(time.Second):
			entryStorage.DeleteExpired()
		case <-stopChan:
			stopChan <- struct{}{}
			close(stopChan)
			return
		}
	}
}

func listen(handlerConfig api.HandlerConfig) *http.Server {
	apiRoot := getAPIRoot(handlerConfig.WebExternalURL)
	fmt.Println("Handle Path: ", apiRoot)

	r := http.NewServeMux()
	r.Handle(apiRoot, http.StripPrefix(apiRoot, api.NewSecretHandler(handlerConfig)))
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

	var entryStorage storage.VerifyStorage
	entryStorage = getStorage(postgresDB)
	if entryStorage == nil {
		return nil, fmt.Errorf("No database backend selected")
	}

	config.EntryStorage = entryStorage

	return &config, nil
}

func main() {
	handlerConfig, err := getConfig()

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s", err)
		os.Exit(1)
	}

	stopChan := make(chan interface{})
	go scheduleDeleteExpired(handlerConfig.EntryStorage, stopChan)
	httpServer := listen(*handlerConfig)

	termChan := make(chan os.Signal)
	signal.Notify(termChan, syscall.SIGTERM, syscall.SIGINT)

	defer close(termChan)
	<-termChan

	ctx := context.Background()
	shutdownErrors := shutDown(func() error {
		ctx, cancel := context.WithTimeout(ctx, time.Second*5)
		defer cancel()
		return httpServer.Shutdown(ctx)
	}, func() error {
		return handlerConfig.EntryStorage.Close()
	}, func() error {
		stopChan <- struct{}{}
		return nil
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
		case _, ok := <-stopChan:
			if !ok {
				stopChan = nil
			}
		case <-time.After(time.Second * 15):
			fmt.Fprint(os.Stderr, "error: force quit")
			os.Exit(2)
		}

		if shutdownErrors == nil && stopChan == nil {
			if errored {
				os.Exit(1)
			}
			return
		}
	}
}
