package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func cleanEntries(t *testing.T) {
	extURL, _ := url.Parse("http://example.com")
	webExternalURL = extURL
	storage = NewMemoryStorage()
}

func TestGetUUIDFromPath(t *testing.T) {
	testCases := []struct {
		Name     string
		Value    string
		Expected string
	}{
		{
			"simple uuid",
			"/3f356f6c-c8b1-4b48-8243-aa04d07b8873",
			"3f356f6c-c8b1-4b48-8243-aa04d07b8873",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			cleanEntries(t)
			actual, err := getUUIDFromPath(testCase.Value)
			if err != nil {
				t.Fatal(err)
			}
			if testCase.Expected != actual {
				t.Errorf("expected: %q, actual: %q", testCase.Expected, actual)
			}
		})
	}
}

func TestCreateEntry(t *testing.T) {
	expireSeconds = 10
	testCases := []string{
		"foo",
	}

	for _, testCase := range testCases {
		t.Run(testCase, func(t *testing.T) {
			cleanEntries(t)
			req := httptest.NewRequest("POST", "http://example.com", bytes.NewReader([]byte(testCase)))
			w := httptest.NewRecorder()
			handleRequest(w, req)

			resp := w.Result()
			body, _ := ioutil.ReadAll(resp.Body)

			responseURL := string(body)
			savedUUID, keyString, err := getUUIDAndSecretFromPath(responseURL)

			if err != nil {
				t.Fatal(err)
			}

			key, err := hex.DecodeString(keyString)

			if err != nil {
				t.Fatal(err)
			}

			secretStorage := &SecretStorage{storage, &AESEncrypter{key}}
			entry, err := secretStorage.Get(savedUUID)

			if err != nil {
				t.Fatal(err)
			}

			actual := string(entry.Data)

			if testCase != actual {
				t.Errorf("data not saved expected: %q, actual: %q", testCase, actual)
			}
		})
	}

}

func TestGetEntry(t *testing.T) {
	testCases := []struct {
		Name  string
		Value string
		UUID  string
	}{
		{
			"first",
			"foo",
			"3f356f6c-c8b1-4b48-8243-aa04d07b8873",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			key, err := generateRSAKey()
			if err != nil {
				t.Fatal(err)
			}
			cleanEntries(t)
			encrypter := AESEncrypter{key}
			encryptedData, err := encrypter.Encrypt([]byte(testCase.Value))
			if err != nil {
				t.Fatal(err)
			}

			storage.Create(testCase.UUID, encryptedData, time.Second*10)

			req := httptest.NewRequest("GET", fmt.Sprintf("http://example.com/%s/%s", testCase.UUID, hex.EncodeToString(key)), nil)
			w := httptest.NewRecorder()
			handleRequest(w, req)

			resp := w.Result()
			body, _ := ioutil.ReadAll(resp.Body)

			actual := string(body)

			if actual != testCase.Value {
				t.Errorf("data not read expected: %q, actual: %q", testCase.Value, actual)
			}
		})
	}

}

func TestSetAndGetEntry(t *testing.T) {
	testCases := []string{
		"foo",
	}

	for _, testCase := range testCases {
		t.Run(testCase, func(t *testing.T) {
			cleanEntries(t)
			req := httptest.NewRequest("POST", "http://example.com", bytes.NewReader([]byte(testCase)))
			w := httptest.NewRecorder()
			handleRequest(w, req)

			resp := w.Result()
			body, _ := ioutil.ReadAll(resp.Body)

			responseURL := string(body)
			savedUUID, keyString, err := getUUIDAndSecretFromPath(responseURL)

			if err != nil {
				t.Fatal(err)
			}

			req = httptest.NewRequest("GET", fmt.Sprintf("http://example.com/%s/%s", savedUUID, keyString), nil)
			w = httptest.NewRecorder()
			handleRequest(w, req)

			resp = w.Result()
			body, _ = ioutil.ReadAll(resp.Body)

			actual := string(body)

			if testCase != actual {
				t.Errorf("data not saved expected: %q, actual: %q", testCase, actual)
			}
		})
	}

}
