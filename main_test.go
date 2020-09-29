package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"
)

func cleanEntries(t *testing.T) {
	extURL, _ := url.Parse("http://example.com")
	webExternalURL = extURL
	entryStorage = newMemoryStorage()
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
	value := "Foo"
	expireSeconds = 10
	cleanEntries(t)
	req := httptest.NewRequest("POST", "http://example.com", bytes.NewReader([]byte(value)))
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

	secretStore := &secretStorage{entryStorage, &AESEncrypter{key}}
	entry, err := secretStore.Get(savedUUID)

	if err != nil {
		t.Fatal(err)
	}

	actual := string(entry.Data)

	if value != actual {
		t.Errorf("data not saved expected: %q, actual: %q", value, actual)
	}
}

func TestCreateEntryJSON(t *testing.T) {
	value := "Foo"
	expireSeconds = 10
	cleanEntries(t)
	req := httptest.NewRequest("POST", "http://example.com", bytes.NewReader([]byte(value)))
	req.Header.Add("Accept", "application/json")
	w := httptest.NewRecorder()
	handleRequest(w, req)

	resp := w.Result()
	var encode SecretResponse
	err := json.NewDecoder(resp.Body).Decode(&encode)

	if err != nil {
		t.Error(err)
	}

	key, err := hex.DecodeString(encode.Key)

	if err != nil {
		t.Fatal(err)
	}

	secretStore := &secretStorage{entryStorage, &AESEncrypter{key}}
	entry, err := secretStore.Get(encode.UUID)

	if err != nil {
		t.Fatal(err)
	}

	actual := string(entry.Data)

	if value != actual {
		t.Errorf("data not saved expected: %q, actual: %q", value, actual)
	}
}

func TestCreateEntryForm(t *testing.T) {
	value := "Foo"
	expireSeconds = 10
	cleanEntries(t)

	data := url.Values{}
	data.Set("secret", value)
	data.Set("expire", "1m")

	req := httptest.NewRequest("POST", "http://example.com", strings.NewReader(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

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

	secretStore := &secretStorage{entryStorage, &AESEncrypter{key}}
	entry, err := secretStore.Get(savedUUID)

	if err != nil {
		t.Fatal(err)
	}

	actual := string(entry.Data)

	if value != actual {
		t.Errorf("data not saved expected: %q, actual: %q", value, actual)
	}

	now := time.Now()
	if entry.Expire.After(now.Add(time.Minute)) {
		t.Errorf("Expiration is more than expected: %q, %q", entry.Expire, now.Add(time.Second*61))
	}
	if entry.Expire.Before(now.Add(time.Second * 59)) {
		t.Errorf("Expiration is less than expected: %q", entry.Expire)
	}
}

func TestRequestPathsCreateEntry(t *testing.T) {
	expireSeconds = 10

	testCases := []struct {
		Name       string
		Path       string
		StatusCode int
	}{
		{Name: "empty path", Path: "", StatusCode: 200},
		{Name: "/ path", Path: "/", StatusCode: 200},
		{Name: "Longer path", Path: "/other", StatusCode: 404},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			cleanEntries(t)
			req := httptest.NewRequest("POST", fmt.Sprintf("http://example.com%s", testCase.Path), bytes.NewReader([]byte("ASDF")))
			w := httptest.NewRecorder()
			handleRequest(w, req)

			resp := w.Result()

			if resp.StatusCode != testCase.StatusCode {
				t.Errorf("Expected statuscode %d, but got %d", testCase.StatusCode, resp.StatusCode)
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

			entryStorage.Create(testCase.UUID, encryptedData, time.Second*10)

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

func TestGetEntryJSON(t *testing.T) {
	testCase := struct {
		Name  string
		Value string
		UUID  string
	}{

		"first",
		"foo",
		"3f356f6c-c8b1-4b48-8243-aa04d07b8873",
	}

	key, err := generateRSAKey()
	if err != nil {
		t.Error(err)
	}
	cleanEntries(t)
	encrypter := AESEncrypter{key}
	encryptedData, err := encrypter.Encrypt([]byte(testCase.Value))
	if err != nil {
		t.Error(err)
	}

	entryStorage.Create(testCase.UUID, encryptedData, time.Second*10)

	req := httptest.NewRequest("GET", fmt.Sprintf("http://example.com/%s/%s", testCase.UUID, hex.EncodeToString(key)), nil)
	req.Header.Add("Accept", "application/json")
	w := httptest.NewRecorder()
	handleRequest(w, req)

	resp := w.Result()
	var encode SecretResponse
	err = json.NewDecoder(resp.Body).Decode(&encode)

	if err != nil {
		t.Error(err)
	}

	if encode.Data != testCase.Value {
		t.Error("Wrong value returned")
	}
}

func TestSetAndGetEntry(t *testing.T) {
	testCase := "foo"

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

func TestCreateEntryWithExpiration(t *testing.T) {
	maxExpireSeconds = 120
	expireSeconds = -10
	testCase := "foo"

	cleanEntries(t)
	req := httptest.NewRequest("POST", "http://example.com?expire=1m", bytes.NewReader([]byte(testCase)))
	w := httptest.NewRecorder()
	handleRequest(w, req)

	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		t.Errorf("Invalid statuscode, expected %d, got %d", 200, resp.StatusCode)
	}

	responseURL := string(body)
	savedUUID, keyString, err := getUUIDAndSecretFromPath(responseURL)

	if err != nil {
		t.Fatal(err)
	}

	key, err := hex.DecodeString(keyString)

	if err != nil {
		t.Fatal(err)
	}

	secretStore := &secretStorage{entryStorage, &AESEncrypter{key}}
	entry, err := secretStore.Get(savedUUID)

	if err != nil {
		t.Fatal(err)
	}

	now := time.Now()
	if entry.Expire.After(now.Add(time.Minute)) {
		t.Errorf("Expiration is more than expected: %q, %q", entry.Expire, now.Add(time.Second*61))
	}
	if entry.Expire.Before(now.Add(time.Second * 59)) {
		t.Errorf("Expiration is less than expected: %q", entry.Expire)
	}
}

func TestCreateEntrySizeLimit(t *testing.T) {
	maxExpireSeconds = 120
	expireSeconds = 10
	testCase := "ff"
	maxDataSize = 1

	cleanEntries(t)
	req := httptest.NewRequest("POST", "http://example.com?expire=1m", bytes.NewReader([]byte(testCase)))
	w := httptest.NewRecorder()
	handleRequest(w, req)

	resp := w.Result()

	if resp.StatusCode != 413 {
		t.Errorf("Invalid statuscode, expected %d, got %d", 413, resp.StatusCode)
	}

}
