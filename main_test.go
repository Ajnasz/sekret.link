package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"
)

func cleanEntries() {
	extURL, _ := url.Parse("http://example.com")
	webExternalURL = extURL
	psqlConnection := newPostgresqlStorage(getPSQLTestConn())
	entryStorage = psqlConnection
	// postgresCleanableStorage{psqlConnection}.Clean()
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
			t.Cleanup(cleanEntries)
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
	t.Cleanup(cleanEntries)
	req := httptest.NewRequest("POST", "http://example.com", bytes.NewReader([]byte(value)))
	w := httptest.NewRecorder()
	secretHandler{}.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	responseURL := string(body)
	savedUUID, keyString, err := getUUIDAndSecretFromPath(responseURL)

	if resp.Header.Get("x-entry-uuid") != savedUUID {
		t.Errorf("Expected x-entry-uuid header to be %q, but got %q", savedUUID, resp.Header.Get("x-entry-uuid"))
	}

	if resp.Header.Get("x-entry-delete-key") == "" {
		t.Error("Expected x-entry-delete-key to not be empty")
	}

	if err != nil {
		t.Fatal(err)
	}

	key, err := hex.DecodeString(keyString)

	if err != nil {
		t.Fatal(err)
	}

	secretStore := &secretStorage{entryStorage, &AESEncrypter{key}}
	entry, err := secretStore.GetAndDelete(savedUUID)

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
	t.Cleanup(cleanEntries)
	req := httptest.NewRequest("POST", "http://example.com", bytes.NewReader([]byte(value)))
	req.Header.Add("Accept", "application/json")
	w := httptest.NewRecorder()
	secretHandler{}.ServeHTTP(w, req)

	resp := w.Result()
	var encode SecretResponse
	err := json.NewDecoder(resp.Body).Decode(&encode)

	if err != nil {
		t.Error(err)
	}

	if encode.DeleteKey == "" {
		t.Error("In create response the deleteKey is empty")
	}

	key, err := hex.DecodeString(encode.Key)

	if err != nil {
		t.Fatal(err)
	}

	secretStore := &secretStorage{entryStorage, &AESEncrypter{key}}
	entry, err := secretStore.GetAndDelete(encode.UUID)

	if err != nil {
		t.Fatal(err)
	}

	actual := string(entry.Data)

	if value != actual {
		t.Errorf("data not saved expected: %q, actual: %q", value, actual)
	}
}

func createMultipart(values map[string]io.Reader) (*bytes.Buffer, *multipart.Writer, error) {
	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	for key, r := range values {
		var fw io.Writer
		var err error
		if x, ok := r.(*os.File); ok {
			if fw, err = w.CreateFormFile(key, x.Name()); err != nil {
				return nil, nil, err
			}
		} else {
			if fw, err = w.CreateFormField(key); err != nil {
				return nil, nil, err
			}
		}

		if _, err = io.Copy(fw, r); err != nil {
			return nil, nil, err
		}
	}

	w.Close()

	return &b, w, nil
}

func TestCreateEntryForm(t *testing.T) {
	value := "Foo"
	expireSeconds = 60
	t.Cleanup(cleanEntries)

	data, multi, err := createMultipart(map[string]io.Reader{
		"secret": strings.NewReader(value),
	})

	req := httptest.NewRequest("POST", fmt.Sprintf("http://example.com/?expire=%ds", expireSeconds), data)
	req.Header.Set("Content-Type", multi.FormDataContentType())

	w := httptest.NewRecorder()

	secretHandler{}.ServeHTTP(w, req)

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
	entry, err := secretStore.GetAndDelete(savedUUID)

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
			t.Cleanup(cleanEntries)
			req := httptest.NewRequest("POST", fmt.Sprintf("http://example.com%s", testCase.Path), bytes.NewReader([]byte("ASDF")))
			w := httptest.NewRecorder()
			secretHandler{}.ServeHTTP(w, req)

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
			t.Cleanup(cleanEntries)
			key, err := generateRSAKey()
			if err != nil {
				t.Fatal(err)
			}
			encrypter := AESEncrypter{key}
			encryptedData, err := encrypter.Encrypt([]byte(testCase.Value))
			if err != nil {
				t.Fatal(err)
			}

			entryStorage.Create(testCase.UUID, encryptedData, time.Second*10, 1)

			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("http://example.com/%s/%s", testCase.UUID, hex.EncodeToString(key)), nil)
			w := httptest.NewRecorder()
			secretHandler{}.ServeHTTP(w, req)

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
	t.Cleanup(cleanEntries)
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
	encrypter := AESEncrypter{key}
	encryptedData, err := encrypter.Encrypt([]byte(testCase.Value))
	if err != nil {
		t.Error(err)
	}

	entryStorage.Create(testCase.UUID, encryptedData, time.Second*10, 1)

	req := httptest.NewRequest("GET", fmt.Sprintf("http://example.com/%s/%s", testCase.UUID, hex.EncodeToString(key)), nil)
	req.Header.Add("Accept", "application/json")
	w := httptest.NewRecorder()
	secretHandler{}.ServeHTTP(w, req)

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

	t.Cleanup(cleanEntries)
	req := httptest.NewRequest("POST", "http://example.com", bytes.NewReader([]byte(testCase)))
	w := httptest.NewRecorder()
	secretHandler{}.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	responseURL := string(body)
	savedUUID, keyString, err := getUUIDAndSecretFromPath(responseURL)

	if err != nil {
		t.Fatal(err)
	}

	req = httptest.NewRequest("GET", fmt.Sprintf("http://example.com/%s/%s", savedUUID, keyString), nil)
	w = httptest.NewRecorder()
	secretHandler{}.ServeHTTP(w, req)

	resp = w.Result()
	body, _ = ioutil.ReadAll(resp.Body)

	actual := string(body)

	if testCase != actual {
		t.Errorf("data not saved expected: %q, actual: %q", testCase, actual)
	}
}

func TestCreateEntryWithExpiration(t *testing.T) {
	t.Cleanup(cleanEntries)
	maxExpireSeconds = 120
	expireSeconds = -10
	testCase := "foo"

	req := httptest.NewRequest("POST", "http://example.com?expire=1m", bytes.NewReader([]byte(testCase)))
	w := httptest.NewRecorder()
	secretHandler{}.ServeHTTP(w, req)

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
	entry, err := secretStore.GetAndDelete(savedUUID)

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
	t.Cleanup(cleanEntries)
	maxExpireSeconds = 120
	expireSeconds = 10
	testCase := "ff"
	oldMaxDataSize := maxDataSize
	defer func() { maxDataSize = oldMaxDataSize }()
	maxDataSize = 1

	req := httptest.NewRequest("POST", "http://example.com?expire=1m", bytes.NewReader([]byte(testCase)))
	w := httptest.NewRecorder()
	secretHandler{}.ServeHTTP(w, req)

	resp := w.Result()

	if resp.StatusCode != 413 {
		t.Errorf("Invalid statuscode, expected %d, got %d", 413, resp.StatusCode)
	}

}

func TestCreateEntryWithMaxReads(t *testing.T) {
	value := "FooBarBAzdd"
	expireSeconds = 10
	t.Cleanup(cleanEntries)
	req := httptest.NewRequest("POST", "http://example.com?maxReads=2", bytes.NewReader([]byte(value)))
	w := httptest.NewRecorder()
	secretHandler{}.ServeHTTP(w, req)

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
	entry, err := secretStore.GetMeta(savedUUID)

	if err != nil {
		t.Fatal(err)
	}

	if entry.MaxReads != 2 {
		t.Errorf("expected max reads to be: %d, actual: %d", 2, entry.MaxReads)
	}
}

func TestDeleteEntry(t *testing.T) {
	t.Cleanup(cleanEntries)
	req := httptest.NewRequest(http.MethodPost, "http://example.com", bytes.NewReader([]byte("foobarbaz")))
	w := httptest.NewRecorder()
	handler := secretHandler{}
	handler.ServeHTTP(w, req)

	resp := w.Result()

	deleteKey := resp.Header.Get("x-entry-delete-key")
	key := resp.Header.Get("x-entry-key")
	UUID := resp.Header.Get("x-entry-uuid")

	url := fmt.Sprintf("http://example.com/%s/%s/%s", UUID, key, deleteKey)

	req = httptest.NewRequest(http.MethodDelete, url, nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	resp = w.Result()

	if resp.StatusCode != http.StatusAccepted {
		t.Errorf("Delete response expected to be %d, but got %d", http.StatusAccepted, resp.StatusCode)
	}
}
