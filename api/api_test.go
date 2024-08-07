package api

import (
	"bytes"
	"context"
	"database/sql"
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

	"github.com/Ajnasz/sekret.link/internal/hasher"
	"github.com/Ajnasz/sekret.link/internal/key"
	"github.com/Ajnasz/sekret.link/internal/models"
	"github.com/Ajnasz/sekret.link/internal/services"
	"github.com/Ajnasz/sekret.link/internal/test/durable"
	"github.com/Ajnasz/sekret.link/internal/uuid"
	"github.com/stretchr/testify/assert"
)

func NewHandlerConfig(db *sql.DB) HandlerConfig {
	extURL, _ := url.Parse("http://example.com")
	return HandlerConfig{
		ExpireSeconds:    10,
		MaxDataSize:      1024 * 1024,
		MaxExpireSeconds: 60 * 60 * 24 * 30,
		WebExternalURL:   extURL,
		DB:               db,
	}
}

func TestCreateEntry(t *testing.T) {
	value := "Foo"
	ctx := context.Background()
	db, err := durable.TestConnection(ctx)

	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	t.Run("happy path", func(t *testing.T) {
		req := httptest.NewRequest("POST", "http://example.com", bytes.NewReader([]byte(value)))
		w := httptest.NewRecorder()
		handlerConfig := NewHandlerConfig(db)
		h := NewSecretHandler(handlerConfig)
		h.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Logf("response: %s", w.Body.String())
			t.Fatalf("expected statuscode to be %d, but got %d", http.StatusOK, w.Code)
		}

		resp := w.Result()
		body, err := io.ReadAll(resp.Body)

		if err != nil {
			t.Fatal(err)
		}

		responseURL := string(body)
		t.Log("responseURL", responseURL)
		savedUUID, keyString, err := uuid.GetUUIDAndSecretFromPath(responseURL)

		assert.NoError(t, err)

		assert.Equal(t, resp.Header.Get("x-entry-uuid"), savedUUID, "wrong x-entry-uuid")

		assert.NotEmpty(t, resp.Header.Get("x-entry-delete-key"), "x-entry-delete-key is empty")

		k, err := key.FromString(keyString)
		assert.NoError(t, err)

		encrypter := func(b key.Key) services.Encrypter {
			return services.NewAESEncrypter(b)
		}
		keyManager := services.NewEntryKeyManager(db, &models.EntryKeyModel{}, hasher.NewSHA256Hasher(), encrypter)

		entryManager := services.NewEntryManager(db, &models.EntryModel{}, encrypter, keyManager)
		entry, err := entryManager.ReadEntry(ctx, savedUUID, *k)

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
			handlerConfig := NewHandlerConfig(db)
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

	ctx := context.Background()
	db, err := durable.TestConnection(ctx)

	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	req := httptest.NewRequest("POST", "http://example.com", bytes.NewReader([]byte(value)))
	req.Header.Add("Accept", "application/json")
	w := httptest.NewRecorder()
	NewSecretHandler(NewHandlerConfig(db)).ServeHTTP(w, req)

	resp := w.Result()

	type SecretResponse struct {
		UUID      string
		Key       string
		Data      string
		Created   time.Time
		Accessed  time.Time
		Expire    time.Time
		DeleteKey string
	}

	var encode SecretResponse
	if err := json.NewDecoder(resp.Body).Decode(&encode); err != nil {
		b, _ := io.ReadAll(resp.Body)
		t.Fatal(err, string(b))
	}

	if encode.DeleteKey == "" {
		t.Error("In create response the deleteKey is empty")
	}

	k, err := key.FromString(encode.Key)

	if err != nil {
		t.Fatal(err)
	}

	encrypter := func(b key.Key) services.Encrypter {
		return services.NewAESEncrypter(b)
	}
	keyManager := services.NewEntryKeyManager(db, &models.EntryKeyModel{}, hasher.NewSHA256Hasher(), encrypter)
	entryManager := services.NewEntryManager(db, &models.EntryModel{}, encrypter, keyManager)
	entry, err := entryManager.ReadEntry(ctx, encode.UUID, *k)

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
	ctx := context.Background()
	db, err := durable.TestConnection(ctx)

	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	data, multi, err := createMultipart(map[string]io.Reader{
		"secret": strings.NewReader(value),
	})

	if err != nil {
		t.Error(err)
	}
	handlerConfig := NewHandlerConfig(db)
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

	k, err := key.FromString(keyString)

	if err != nil {
		t.Fatal(err)
	}

	encrypter := func(b key.Key) services.Encrypter {
		return services.NewAESEncrypter(b)
	}
	keyManager := services.NewEntryKeyManager(db, &models.EntryKeyModel{}, hasher.NewSHA256Hasher(), encrypter)
	entryManager := services.NewEntryManager(db, &models.EntryModel{}, encrypter, keyManager)
	entry, err := entryManager.ReadEntry(ctx, savedUUID, *k)

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

	ctx := context.Background()
	db, err := durable.TestConnection(ctx)

	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			req := httptest.NewRequest("POST", fmt.Sprintf("http://example.com%s", testCase.Path), bytes.NewReader([]byte("ASDF")))
			w := httptest.NewRecorder()
			NewSecretHandler(NewHandlerConfig(db)).ServeHTTP(w, req)

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
	}{
		{
			"first",
			"foo",
		},
	}

	ctx := context.Background()
	db, err := durable.TestConnection(ctx)

	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			encrypter := func(b key.Key) services.Encrypter {
				return services.NewAESEncrypter(b)
			}

			keyManager := services.NewEntryKeyManager(db, &models.EntryKeyModel{}, hasher.NewSHA256Hasher(), encrypter)
			entryManager := services.NewEntryManager(db, &models.EntryModel{}, encrypter, keyManager)
			expire := time.Second * 10
			maxReads := 1
			meta, encKey, err := entryManager.CreateEntry(ctx, "text/plain", []byte(testCase.Value), &expire, &maxReads)

			if err != nil {
				t.Fatal(err)
			}

			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("http://example.com/%s/%s", meta.UUID, encKey.String()), nil)
			w := httptest.NewRecorder()

			mux := http.NewServeMux()
			secretHandler := NewSecretHandler(NewHandlerConfig(db))
			secretHandler.RegisterHandlers(mux, "")

			mux.ServeHTTP(w, req)

			resp := w.Result()

			if resp.StatusCode != http.StatusOK {
				t.Fatalf("expected statuscode %d got %d", http.StatusOK, resp.StatusCode)
			}
			body, _ := io.ReadAll(resp.Body)

			actual := string(body)

			if actual != testCase.Value {
				t.Errorf("data not read expected: %q, actual: %q", testCase.Value, actual)
			}
		})
	}
}

func TestGetEntryJSON(t *testing.T) {
	ctx := context.Background()
	db, err := durable.TestConnection(ctx)

	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		defer db.Close()
	})
	testCase := struct {
		Name  string
		Value string
	}{

		"first",
		"foo",
	}

	encrypter := func(b key.Key) services.Encrypter {
		return services.NewAESEncrypter(b)
	}

	keyManager := services.NewEntryKeyManager(db, &models.EntryKeyModel{}, hasher.NewSHA256Hasher(), encrypter)
	entryManager := services.NewEntryManager(db, &models.EntryModel{}, encrypter, keyManager)
	expire := time.Second * 10
	maxReads := 1
	meta, encKey, err := entryManager.CreateEntry(ctx, "text/plain", []byte(testCase.Value), &expire, &maxReads)
	if err != nil {
		t.Error(err)
	}

	req := httptest.NewRequest("GET", fmt.Sprintf("http://example.com/%s/%s", meta.UUID, encKey.String()), nil)
	req.Header.Add("Accept", "application/json")
	w := httptest.NewRecorder()

	mux := http.NewServeMux()
	secretHandler := NewSecretHandler(NewHandlerConfig(db))
	secretHandler.RegisterHandlers(mux, "")
	mux.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != 200 {
		t.Errorf("non 200 http statuscode: %d", resp.StatusCode)
	}

	type SecretResponse struct {
		UUID      string
		Key       string
		Data      string
		Created   time.Time
		Accessed  time.Time
		Expire    time.Time
		DeleteKey string
	}

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

	ctx := context.Background()
	db, err := durable.TestConnection(ctx)

	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	req := httptest.NewRequest("POST", "http://example.com/", bytes.NewReader([]byte(testCase)))
	w := httptest.NewRecorder()

	mux := http.NewServeMux()
	secretHandler := NewSecretHandler(NewHandlerConfig(db))
	secretHandler.RegisterHandlers(mux, "")
	mux.ServeHTTP(w, req)

	resp := w.Result()

	if resp.StatusCode != 200 {
		t.Fatalf("expected statuscode %d got %d", 200, resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)

	responseURL := string(body)
	savedUUID, keyString, err := uuid.GetUUIDAndSecretFromPath(responseURL)

	if err != nil {
		t.Fatal(err)
	}

	req = httptest.NewRequest("GET", fmt.Sprintf("http://example.com/%s/%s", savedUUID, keyString), nil)
	w = httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	resp = w.Result()

	assert.Equal(t, 200, resp.StatusCode, "bad status code")
	body, _ = io.ReadAll(resp.Body)

	actual := string(body)

	assert.Equal(t, testCase, actual, "data not saved")
}

func TestCreateEntryWithExpiration(t *testing.T) {
	ctx := context.Background()
	db, err := durable.TestConnection(ctx)

	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		db.Close()
	})
	testCase := "foo"

	req := httptest.NewRequest("POST", "http://example.com?expire=1m", bytes.NewReader([]byte(testCase)))
	w := httptest.NewRecorder()
	handlerConfig := NewHandlerConfig(db)
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

	decodedKey, err := key.FromString(keyString)

	if err != nil {
		t.Fatal(err)
	}

	encrypter := func(b key.Key) services.Encrypter {
		return services.NewAESEncrypter(b)
	}
	keyManager := services.NewEntryKeyManager(db, &models.EntryKeyModel{}, hasher.NewSHA256Hasher(), encrypter)
	entryManager := services.NewEntryManager(db, &models.EntryModel{}, encrypter, keyManager)
	entry, err := entryManager.ReadEntry(ctx, savedUUID, *decodedKey)

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
	ctx := context.Background()
	db, err := durable.TestConnection(ctx)

	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		db.Close()
	})
	testCase := "ff"

	req := httptest.NewRequest("POST", "http://example.com?expire=1m", bytes.NewReader([]byte(testCase)))
	w := httptest.NewRecorder()
	handlerConfig := NewHandlerConfig(db)
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
	ctx := context.Background()
	db, err := durable.TestConnection(ctx)

	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	req := httptest.NewRequest("POST", "http://example.com?maxReads=2", bytes.NewReader([]byte(value)))
	w := httptest.NewRecorder()
	NewSecretHandler(NewHandlerConfig(db)).ServeHTTP(w, req)

	resp := w.Result()

	savedUUID := resp.Header.Get("x-entry-uuid")

	if err != nil {
		t.Fatal(err)
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	model := &models.EntryKeyModel{}
	entries, err := model.Get(ctx, tx, savedUUID)

	if err != nil {
		if err := tx.Rollback(); err != nil {
			t.Error(err)
		}
		t.Fatal(err)
	}

	if err := tx.Commit(); err != nil {
		t.Errorf("commit failed: %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("expected to get entry key %d, got %d", 1, len(entries))
	}

	remainingReads := entries[0].RemainingReads.Int16
	if remainingReads != 2 {
		t.Fatalf("expected max reads to be: %d, actual: %d", 2, remainingReads)
	}
}

func Test_DeleteEntry(t *testing.T) {
	ctx := context.Background()
	db, err := durable.TestConnection(ctx)

	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	t.Run("correct key", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "http://example.com", bytes.NewReader([]byte("foobarbaz")))
		w := httptest.NewRecorder()
		handler := NewSecretHandler(NewHandlerConfig(db))
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
		handler := NewSecretHandler(NewHandlerConfig(db))
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
			t.Errorf("Delete response expected to be %d or %d, but got %d", http.StatusUnauthorized, http.StatusNotFound, resp.StatusCode)
		}
	})
	t.Run("without delete key", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "http://example.com", bytes.NewReader([]byte("foobarbaz")))
		w := httptest.NewRecorder()
		handler := NewSecretHandler(NewHandlerConfig(db))
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
	ctx := context.Background()
	db, err := durable.TestConnection(ctx)

	if err != nil {
		f.Fatal(err)
	}

	f.Cleanup(func() {
		defer db.Close()
	})

	f.Fuzz(func(t *testing.T, testCase string) {
		if testCase == "" {
			t.Log("empty")
			return
		}

		mux := http.NewServeMux()
		secretHandler := NewSecretHandler(NewHandlerConfig(db))
		secretHandler.RegisterHandlers(mux, "")

		req := httptest.NewRequest("POST", "http://example.com/", bytes.NewReader([]byte(testCase)))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		resp := w.Result()
		body, _ := io.ReadAll(resp.Body)

		responseURL := string(body)
		savedUUID, keyString, err := uuid.GetUUIDAndSecretFromPath(responseURL)

		if err != nil {
			t.Fatal(err)
		}

		req = httptest.NewRequest("GET", fmt.Sprintf("http://example.com/%s/%s", savedUUID, keyString), nil)
		w = httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		resp = w.Result()
		body, _ = io.ReadAll(resp.Body)

		actual := string(body)

		if testCase != actual {
			t.Errorf("data not saved expected: %q, actual: %q", testCase, actual)
		}
	})
}
