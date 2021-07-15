package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/Ajnasz/sekret.link/config"
	"github.com/Ajnasz/sekret.link/storage"
)

var (
	entryStorage     storage.VerifyStorage
	externalURLParam string
	expireSeconds    int
	maxExpireSeconds int
	postgresDB       string
	webExternalURL   *url.URL
	maxDataSize      int64
	version          string
	queryVersion     bool
)

func getStorage() storage.VerifyStorage {
	return storage.NewStorage(config.GetConnectionString(postgresDB))
}

func init() {
	flag.StringVar(&externalURLParam, "webExternalURL", "", "Web server external url")
	flag.StringVar(&postgresDB, "postgresDB", "", "Connection string for postgresql database backend")
	flag.IntVar(&expireSeconds, "expireSeconds", 60*60*24*7, "Default expiration time in seconds")
	flag.IntVar(&maxExpireSeconds, "maxExpireSeconds", 60*60*24*30, "Max expiration time in seconds")
	flag.Int64Var(&maxDataSize, "maxDataSize", 1024*1024, "Max data size")
	flag.BoolVar(&queryVersion, "version", false, "Get version information")
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

func scheduleDeleteExpired(stopChan chan interface{}) {
	for {
		select {
		case <-time.After(time.Second):
			entryStorage.DeleteExpired()
		case <-stopChan:
			return
		}
	}
}

func listen(apiRoot string) *http.Server {
	log.Println("Handle Path: ", apiRoot)

	r := http.NewServeMux()
	r.Handle(apiRoot, http.StripPrefix(apiRoot, secretHandler{}))
	httpServer := &http.Server{
		Addr:         ":8080",
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		if err := httpServer.ListenAndServe(); err != nil {
			if err.Error() != "http: Server closed" {
				log.Fatal(err)
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

func main() {
	flag.Parse()

	if queryVersion {
		fmt.Println(version)
		return
	}

	if maxExpireSeconds < expireSeconds {
		log.Fatal("`expireSeconds` must be less or equal then `maxExpireSeconds`")
	}

	extURL, err := url.Parse(externalURLParam)

	if err != nil {
		log.Fatal(err)
	}

	entryStorage = getStorage()
	if entryStorage == nil {
		log.Fatal("No database backend selected")
	}

	stopChan := make(chan interface{})
	go scheduleDeleteExpired(stopChan)

	webExternalURL = extURL
	httpServer := listen(getAPIRoot(extURL))

	termChan := make(chan os.Signal)
	signal.Notify(termChan, syscall.SIGTERM, syscall.SIGINT)

	defer close(termChan)
	<-termChan
	ctx := context.Background()
	// on close
	c := shutDown(func() error {
		ctx, cancel := context.WithTimeout(ctx, time.Second*5)
		defer cancel()
		return httpServer.Shutdown(ctx)
	}, func() error {
		return entryStorage.Close()
	}, func() error {
		stopChan <- struct{}{}
		return nil
	})

	for err := range c {
		if err != nil {
			log.Fatal(err)
		}
	}

	os.Exit(0)
}
