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
		log.Println(err)
		return defaultExpire, nil
	}

	maxExpire := time.Duration(maxExpireSeconds) * time.Second

	if userExpire > maxExpire {
		return 0, fmt.Errorf("Invalid expiration date")
	}

	return userExpire, nil
}

func handleCreateEntry(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)

	if err != nil {
		log.Println(err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	key, keyString, err := createKey()

	if err != nil {
		log.Println(err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	secretStorage := &SecretStorage{storage, &AESEncrypter{key}}

	UUID := newUUIDString()

	expiration, err := getExpiration(r.URL.Query().Get("expire"), time.Second*time.Duration(expireSeconds))

	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid expiration date", http.StatusBadRequest)
		return
	}

	err = secretStorage.Create(UUID, body, expiration)

	if err != nil {
		log.Println(err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Add("x-entry-uuid", UUID)
	w.Header().Add("x-entry-key", keyString)
	if r.Header.Get("Accept") == "application/json" {
		w.Header().Set("Content-Type", "application/json")
		entry, err := secretStorage.GetMeta(UUID)

		if err != nil {
			log.Println(err)
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}

		response := secretResponseFromEntryMeta(entry)
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
