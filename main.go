package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

var (
	entryStorage     verifyStorage
	externalURLParam string
	expireSeconds    int
	maxExpireSeconds int
	postgresDB       string
	webExternalURL   *url.URL
	maxDataSize      int64
	version          string
	queryVersion     bool
)

func getStorage() verifyStorage {
	return newStorage()
}

func init() {
	flag.StringVar(&externalURLParam, "webExternalURL", "", "Web server external url")
	flag.StringVar(&postgresDB, "postgresDB", "", "Connection string for postgresql database backend")
	flag.IntVar(&expireSeconds, "expireSeconds", 60*60*24*7, "Default expiration time in seconds")
	flag.IntVar(&maxExpireSeconds, "maxExpireSeconds", 60*60*24*30, "Max expiration time in seconds")
	flag.Int64Var(&maxDataSize, "maxDataSize", 1024*1024, "Max data size")
	flag.BoolVar(&queryVersion, "version", false, "Get version information")
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

	webExternalURL = extURL

	entryStorage = getStorage()
	if entryStorage == nil {
		log.Fatal("No database backend selected")
	}

	stopChan := make(chan interface{})
	go func() {
		for {
			select {
			case <-time.After(time.Second):
				entryStorage.DeleteExpired()
			case <-stopChan:
				return
			}
		}
	}()

	defer entryStorage.Close()
	defer func() { stopChan <- struct{}{} }()

	apiRoot := ""

	if webExternalURL.Path != "" {
		apiRoot = webExternalURL.Path
	}

	apiRoot = path.Clean(apiRoot)

	if !strings.HasSuffix(apiRoot, "/") {
		apiRoot += "/"
	}

	log.Println("Handle Path: ", apiRoot)

	http.Handle(apiRoot, http.StripPrefix(apiRoot, secretHandler{}))

	log.Fatal(http.ListenAndServe(":8080", nil))
}
