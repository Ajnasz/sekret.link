package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

func createKey() ([]byte, string, error) {
	key, err := generateRSAKey()
	if err != nil {
		return nil, "", err
	}

	keyString := hex.EncodeToString(key)

	return key, keyString, nil
}

func getExpiration(expire string, defaultExpire time.Duration) (time.Duration, error) {
	if expire == "" {
		return defaultExpire, nil
	}
	userExpire, err := time.ParseDuration(expire)
	if err != nil {
		return 0, err
	}

	maxExpire := time.Duration(maxExpireSeconds) * time.Second

	if userExpire > maxExpire {
		return 0, fmt.Errorf("Invalid expiration date")
	}

	return userExpire, nil
}

func getRequestBody(r *http.Request) ([]byte, error) {
	var body []byte
	var err error
	if r.Header.Get("Content-Type") == "application/x-www-form-urlencoded" {
		if r.PostForm == nil {
			err = r.ParseForm()
			if err != nil {
				return nil, err
			}
		}
		body = []byte(r.Form.Get("secret"))
	} else {
		body, err = ioutil.ReadAll(r.Body)
	}

	return body, err
}

func getExpirationR(r *http.Request) (time.Duration, error) {
	var expiration string
	if r.Header.Get("Content-Type") == "application/x-www-form-urlencoded" {
		if r.PostForm == nil {
			err := r.ParseForm()
			if err != nil {
				return 0, err
			}
		}

		expiration = r.Form.Get("expire")
	} else {
		expiration = r.URL.Query().Get("expire")
	}

	return getExpiration(expiration, time.Second*time.Duration(expireSeconds))
}

func handleCreateEntry(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxDataSize)

	body, err := getRequestBody(r)

	if err != nil {
		log.Println(err)
		if err.Error() == "http: request body too large" {
			http.Error(w, "Too large", http.StatusRequestEntityTooLarge)
		} else {
			http.Error(w, "Internal error", http.StatusInternalServerError)
		}
		return
	}

	key, keyString, err := createKey()

	if err != nil {
		log.Println(err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	secretStore := &secretStorage{entryStorage, &AESEncrypter{key}}

	UUID := newUUIDString()

	expiration, err := getExpirationR(r)

	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid expiration", http.StatusBadRequest)
		return
	}

	err = secretStore.Create(UUID, body, expiration)

	if err != nil {
		log.Println(err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Add("x-entry-uuid", UUID)
	w.Header().Add("x-entry-key", keyString)
	w.Header().Add("x-entry-expire", time.Now().Add(expiration).Format(time.RFC3339))
	if r.Header.Get("Accept") == "application/json" {
		w.Header().Set("Content-Type", "application/json")
		entry, err := secretStore.GetMeta(UUID)

		if err != nil {
			log.Println(err)
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}

		response := secretResponseFromEntryMeta(*entry)
		response.Key = keyString

		json.NewEncoder(w).Encode(response)
	} else {
		newURL, err := getUUIDUrlWithSecret(webExternalURL, UUID, keyString)
		if err != nil {
			log.Println(err)
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "%s", newURL.String())
	}
}
