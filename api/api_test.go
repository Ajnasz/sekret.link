package api

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Ajnasz/sekret.link/api/entries"
	"github.com/Ajnasz/sekret.link/encrypter/aes"
	"github.com/Ajnasz/sekret.link/key"
	"github.com/Ajnasz/sekret.link/storage"
	"github.com/Ajnasz/sekret.link/storage/postgresql"
	"github.com/Ajnasz/sekret.link/storage/secret"
	"github.com/Ajnasz/sekret.link/testhelper"
	"github.com/Ajnasz/sekret.link/uuid"
)

func NewHandlerConfig(storage storage.VerifyConfirmReader) HandlerConfig {
	extURL, _ := url.Parse("http://example.com")
	return HandlerConfig{
		ExpireSeconds:    10,
		EntryStorage:     storage,
		MaxDataSize:      1024 * 1024,
		MaxExpireSeconds: 60 * 60 * 24 * 30,
		WebExternalURL:   extURL,
	}
}

func TestCreateEntry(t *testing.T) {
	value := "Foo"
	connection := postgresql.NewStorage(testhelper.GetPSQLTestConn())
	t.Cleanup(func() {
		connection.Close()
	})

	t.Run("happy path", func(t *testing.T) {
		req := httptest.NewRequest("POST", "http://example.com", bytes.NewReader([]byte(value)))
		w := httptest.NewRecorder()
		handlerConfig := NewHandlerConfig(connection)
		h := NewSecretHandler(handlerConfig)
		h.ServeHTTP(w, req)

		resp := w.Result()
		body, _ := io.ReadAll(resp.Body)

		responseURL := string(body)
		savedUUID, keyString, err := uuid.GetUUIDAndSecretFromPath(responseURL)

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

		secretStore := secret.NewSecretStorage(connection, aes.New(key))
		ctx := context.Background()
		entry, err := secretStore.Read(ctx, savedUUID)

		if err != nil {
			t.Fatal(err)
		}

		actual := string(entry.Data)

		if value != actual {
			t.Errorf("data not saved expected: %q, actual: %q", value, actual)
		}
	})
	t.Run("sad path", func(t *testing.T) {
		testCases := []struct {
			qs         string
			statusCode int
			message    string
			body       string
		}{
			{
				qs:         "expire=-1s",
				statusCode: http.StatusBadRequest,
				message:    "Invalid expiration",
				body:       "test",
			},
			{
				qs:         "expire=121s",
				statusCode: http.StatusBadRequest,
				message:    "Invalid expiration",
				body:       "test",
			},
			{
				qs:         "maxReads=0",
				statusCode: http.StatusBadRequest,
				message:    "Invalid max read",
				body:       "test",
			},
			{
				qs:         "maxReads=abc",
				statusCode: http.StatusBadRequest,
				message:    "Invalid max read",
				body:       "test",
			},
			{
				qs:         "",
				statusCode: http.StatusBadRequest,
				message:    "Invalid data",
				body:       "",
			},
		}
		for _, testCase := range testCases {
			req := httptest.NewRequest("POST", "http://example.com?"+testCase.qs, bytes.NewReader([]byte(testCase.body)))
			w := httptest.NewRecorder()
			handlerConfig := NewHandlerConfig(connection)
			handlerConfig.MaxExpireSeconds = 120
			h := NewSecretHandler(handlerConfig)
			h.ServeHTTP(w, req)

			resp := w.Result()

			if resp.StatusCode != testCase.statusCode {
				t.Errorf("expected statuscode to be %d, got %d", testCase.statusCode, resp.StatusCode)
			}
		}
	})
}

func TestCreateEntryJSON(t *testing.T) {
	value := "Foo"
	connection := postgresql.NewStorage(testhelper.GetPSQLTestConn())
	t.Cleanup(func() {
		connection.Close()
	})
	req := httptest.NewRequest("POST", "http://example.com", bytes.NewReader([]byte(value)))
	req.Header.Add("Accept", "application/json")
	w := httptest.NewRecorder()
	NewSecretHandler(NewHandlerConfig(connection)).ServeHTTP(w, req)

	resp := w.Result()
	var encode entries.SecretResponse
	err := json.NewDecoder(resp.Body).Decode(&encode)

	if err != nil {
		b, _ := io.ReadAll(resp.Body)
		t.Error(err, string(b))
	}

	if encode.DeleteKey == "" {
		t.Error("In create response the deleteKey is empty")
	}

	key, err := hex.DecodeString(encode.Key)

	if err != nil {
		t.Fatal(err)
	}

	secretStore := secret.NewSecretStorage(connection, aes.New(key))
	ctx := context.Background()
	entry, err := secretStore.Read(ctx, encode.UUID)

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
	connection := postgresql.NewStorage(testhelper.GetPSQLTestConn())
	t.Cleanup(func() {
		defer connection.Close()
	})

	data, multi, err := createMultipart(map[string]io.Reader{
		"secret": strings.NewReader(value),
	})

	if err != nil {
		t.Error(err)
	}
	handlerConfig := NewHandlerConfig(connection)
	handlerConfig.ExpireSeconds = 60
	req := httptest.NewRequest("POST", fmt.Sprintf("http://example.com/?expire=%ds", handlerConfig.ExpireSeconds), data)
	req.Header.Set("Content-Type", multi.FormDataContentType())

	w := httptest.NewRecorder()

	NewSecretHandler(handlerConfig).ServeHTTP(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	responseURL := string(body)
	savedUUID, keyString, err := uuid.GetUUIDAndSecretFromPath(responseURL)

	if err != nil {
		t.Log("Get UUID And Secret From Path err", err, responseURL)
		t.Fatal(err)
	}

	key, err := hex.DecodeString(keyString)

	if err != nil {
		t.Fatal(err)
	}

	secretStore := secret.NewSecretStorage(connection, aes.New(key))
	ctx := context.Background()
	entry, err := secretStore.Read(ctx, savedUUID)

	if err != nil {
		t.Fatal("Getting entry", err)
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
	testCases := []struct {
		Name       string
		Path       string
		StatusCode int
	}{
		{Name: "empty path", Path: "", StatusCode: 200},
		{Name: "/ path", Path: "/", StatusCode: 200},
		{Name: "Longer path", Path: "/other", StatusCode: 404},
	}
	connection := postgresql.NewStorage(testhelper.GetPSQLTestConn())
	t.Cleanup(func() {
		connection.Close()
	})

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			req := httptest.NewRequest("POST", fmt.Sprintf("http://example.com%s", testCase.Path), bytes.NewReader([]byte("ASDF")))
			w := httptest.NewRecorder()
			NewSecretHandler(NewHandlerConfig(connection)).ServeHTTP(w, req)

			resp := w.Result()

			if resp.StatusCode != testCase.StatusCode {
				body, _ := io.ReadAll(resp.Body)
				t.Errorf("Expected statuscode %d, but got %d, err %s", testCase.StatusCode, resp.StatusCode, body)
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
			uuid.NewUUIDString(),
		},
	}

	connection := postgresql.NewStorage(testhelper.GetPSQLTestConn())
	t.Cleanup(func() {
		connection.Close()
	})

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			k := key.NewKey()
			if err := k.Generate(); err != nil {
				t.Error(err)
			}
			rsakey := k.Get()
			encrypter := aes.New(rsakey)
			encryptedData, err := encrypter.Encrypt([]byte(testCase.Value))
			if err != nil {
				t.Fatal(err)
			}

			ctx := context.Background()
			connection.Write(ctx, testCase.UUID, encryptedData, time.Second*10, 1)

			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("http://example.com/%s/%s", testCase.UUID, hex.EncodeToString(rsakey)), nil)
			w := httptest.NewRecorder()
			NewSecretHandler(NewHandlerConfig(connection)).ServeHTTP(w, req)

			resp := w.Result()
			body, _ := io.ReadAll(resp.Body)

			actual := string(body)

			if actual != testCase.Value {
				t.Errorf("data not read expected: %q, actual: %q", testCase.Value, actual)
			}
		})
	}
}

func TestGetEntryJSON(t *testing.T) {
	connection := postgresql.NewStorage(testhelper.GetPSQLTestConn())
	t.Cleanup(func() {
		defer connection.Close()
	})
	testCase := struct {
		Name  string
		Value string
		UUID  string
	}{

		"first",
		"foo",
		uuid.NewUUIDString(),
	}

	k := key.NewKey()
	if err := k.Generate(); err != nil {
		t.Error(err)
	}
	rsakey := k.Get()

	encrypter := aes.New(rsakey)
	encryptedData, err := encrypter.Encrypt([]byte(testCase.Value))
	if err != nil {
		t.Error(err)
	}

	ctx := context.Background()
	if _, err := connection.Write(ctx, testCase.UUID, encryptedData, time.Second*10, 1); err != nil {
		t.Error(err)
	}

	req := httptest.NewRequest("GET", fmt.Sprintf("http://example.com/%s/%s", testCase.UUID, hex.EncodeToString(rsakey)), nil)
	req.Header.Add("Accept", "application/json")
	w := httptest.NewRecorder()
	NewSecretHandler(NewHandlerConfig(connection)).ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != 200 {
		t.Errorf("non 200 http statuscode: %d", resp.StatusCode)
	}
	var encode entries.SecretResponse
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

	connection := postgresql.NewStorage(testhelper.GetPSQLTestConn())
	t.Cleanup(func() {
		connection.Close()
	})

	req := httptest.NewRequest("POST", "http://example.com", bytes.NewReader([]byte(testCase)))
	w := httptest.NewRecorder()

	NewSecretHandler(NewHandlerConfig(connection)).ServeHTTP(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	responseURL := string(body)
	savedUUID, keyString, err := uuid.GetUUIDAndSecretFromPath(responseURL)

	if err != nil {
		t.Fatal(err)
	}

	req = httptest.NewRequest("GET", fmt.Sprintf("http://example.com/%s/%s", savedUUID, keyString), nil)
	w = httptest.NewRecorder()
	NewSecretHandler(NewHandlerConfig(connection)).ServeHTTP(w, req)

	resp = w.Result()
	body, _ = io.ReadAll(resp.Body)

	actual := string(body)

	if testCase != actual {
		t.Errorf("data not saved expected: %q, actual: %q", testCase, actual)
	}
}

func TestCreateEntryWithExpiration(t *testing.T) {
	connection := postgresql.NewStorage(testhelper.GetPSQLTestConn())
	t.Cleanup(func() {
		connection.Close()
	})
	testCase := "foo"

	req := httptest.NewRequest("POST", "http://example.com?expire=1m", bytes.NewReader([]byte(testCase)))
	w := httptest.NewRecorder()
	handlerConfig := NewHandlerConfig(connection)
	handlerConfig.MaxExpireSeconds = 120
	NewSecretHandler(handlerConfig).ServeHTTP(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		t.Errorf("Invalid statuscode, expected %d, got %d", 200, resp.StatusCode)
	}

	responseURL := string(body)
	savedUUID, keyString, err := uuid.GetUUIDAndSecretFromPath(responseURL)

	if err != nil {
		t.Fatal(err)
	}

	decodedKey, err := hex.DecodeString(keyString)

	if err != nil {
		t.Fatal(err)
	}

	secretStore := secret.NewSecretStorage(connection, aes.New(decodedKey))
	ctx := context.Background()
	entry, err := secretStore.Read(ctx, savedUUID)

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
	connection := postgresql.NewStorage(testhelper.GetPSQLTestConn())
	t.Cleanup(func() {
		connection.Close()
	})
	testCase := "ff"

	req := httptest.NewRequest("POST", "http://example.com?expire=1m", bytes.NewReader([]byte(testCase)))
	w := httptest.NewRecorder()
	handlerConfig := NewHandlerConfig(connection)
	handlerConfig.MaxDataSize = 1
	handlerConfig.MaxExpireSeconds = 120
	NewSecretHandler(handlerConfig).ServeHTTP(w, req)

	resp := w.Result()

	if resp.StatusCode != 413 {
		t.Errorf("Invalid statuscode, expected %d, got %d", 413, resp.StatusCode)
	}

}

func TestCreateEntryWithMaxReads(t *testing.T) {
	value := "FooBarBAzdd"
	connection := postgresql.NewStorage(testhelper.GetPSQLTestConn())
	t.Cleanup(func() {
		connection.Close()
	})

	req := httptest.NewRequest("POST", "http://example.com?maxReads=2", bytes.NewReader([]byte(value)))
	w := httptest.NewRecorder()
	NewSecretHandler(NewHandlerConfig(connection)).ServeHTTP(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	responseURL := string(body)
	savedUUID, keyString, err := uuid.GetUUIDAndSecretFromPath(responseURL)

	if err != nil {
		t.Fatal(err)
	}

	decodedKey, err := hex.DecodeString(keyString)

	if err != nil {
		t.Fatal(err)
	}

	secretStore := secret.NewSecretStorage(connection, aes.New(decodedKey))
	ctx := context.Background()
	entry, err := secretStore.Read(ctx, savedUUID)

	if err != nil {
		t.Fatal(err)
	}

	if entry.MaxReads != 2 {
		t.Errorf("expected max reads to be: %d, actual: %d", 2, entry.MaxReads)
	}
}

func Test_DeleteEntry(t *testing.T) {
	connection := postgresql.NewStorage(testhelper.GetPSQLTestConn())
	t.Cleanup(func() {
		connection.Close()
	})

	t.Run("correct key", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "http://example.com", bytes.NewReader([]byte("foobarbaz")))
		w := httptest.NewRecorder()
		handler := NewSecretHandler(NewHandlerConfig(connection))
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
	})

	t.Run("invalid key", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "http://example.com", bytes.NewReader([]byte("foobarbaz")))
		w := httptest.NewRecorder()
		handler := NewSecretHandler(NewHandlerConfig(connection))
		handler.ServeHTTP(w, req)

		resp := w.Result()

		deleteKey := resp.Header.Get("x-entry-delete-key")
		key := resp.Header.Get("x-entry-key")
		UUID := resp.Header.Get("x-entry-uuid")

		url := fmt.Sprintf("http://example.com/%s/%s/%s", UUID, key, deleteKey+"asdf")

		req = httptest.NewRequest(http.MethodDelete, url, nil)
		w = httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		resp = w.Result()

		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("Delete response expected to be %d, but got %d", http.StatusUnauthorized, resp.StatusCode)
		}
	})
	t.Run("without delete key", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "http://example.com", bytes.NewReader([]byte("foobarbaz")))
		w := httptest.NewRecorder()
		handler := NewSecretHandler(NewHandlerConfig(connection))
		handler.ServeHTTP(w, req)

		resp := w.Result()

		// deleteKey := resp.Header.Get("x-entry-delete-key")
		key := resp.Header.Get("x-entry-key")
		UUID := resp.Header.Get("x-entry-uuid")

		url := fmt.Sprintf("http://example.com/%s/%s", UUID, key)

		req = httptest.NewRequest(http.MethodDelete, url, nil)
		w = httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		resp = w.Result()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Delete response expected to be %d, but got %d", http.StatusBadRequest, resp.StatusCode)
		}
	})
}

func FuzzSetAndGetEntry(f *testing.F) {
	testCases := []string{"foo", " ", "12345", "A3@3!$", string("\xf2")}
	for _, tc := range testCases {
		f.Add(tc) // Use f.Add to provide a seed corpus
	}
	connection := postgresql.NewStorage(testhelper.GetPSQLTestConn())
	f.Cleanup(func() {
		defer connection.Close()
	})

	f.Fuzz(func(t *testing.T, testCase string) {
		if testCase == "" {
			t.Log("empty")
			return
		}
		req := httptest.NewRequest("POST", "http://example.com", bytes.NewReader([]byte(testCase)))
		w := httptest.NewRecorder()

		NewSecretHandler(NewHandlerConfig(connection)).ServeHTTP(w, req)

		resp := w.Result()
		body, _ := io.ReadAll(resp.Body)

		responseURL := string(body)
		savedUUID, keyString, err := uuid.GetUUIDAndSecretFromPath(responseURL)

		if err != nil {
			t.Fatal(err)
		}

		req = httptest.NewRequest("GET", fmt.Sprintf("http://example.com/%s/%s", savedUUID, keyString), nil)
		w = httptest.NewRecorder()
		NewSecretHandler(NewHandlerConfig(connection)).ServeHTTP(w, req)

		resp = w.Result()
		body, _ = io.ReadAll(resp.Body)

		actual := string(body)

		if testCase != actual {
			t.Errorf("data not saved expected: %q, actual: %q", testCase, actual)
		}
	})
}
