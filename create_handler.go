package main

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

func handleCreateEntry(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)

	if err != nil {
		log.Println(err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	UUID := newUUIDString()

	key, err := generateRSAKey()
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	keyString := hex.EncodeToString(key)

	newURL, err := getUUIDUrlWithSecret(webExternalURL, UUID, keyString)

	if err != nil {
		log.Println(err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	secretStorage := &SecretStorage{storage, &AESEncrypter{key}}

	err = secretStorage.Create(UUID, body)

	if err != nil {
		log.Println(err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Add("x-entry-uuid", UUID)
	w.Header().Add("x-entry-key", keyString)
	fmt.Fprintf(w, "%s", newURL.String())
}
