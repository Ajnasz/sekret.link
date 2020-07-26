package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http/httptest"
	"sync"
	"testing"
)

func cleanEntries(t *testing.T) {
	entries = struct {
		sync.RWMutex
		m map[string][]byte
	}{m: make(map[string][]byte)}
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
	testCases := []string{
		"foo",
	}

	for _, testCase := range testCases {
		req := httptest.NewRequest("POST", "http://example.com", bytes.NewReader([]byte(testCase)))
		w := httptest.NewRecorder()
		createHandler(w, req)

		resp := w.Result()
		body, _ := ioutil.ReadAll(resp.Body)

		savedUUID := string(body)
		actual := string(entries.m[savedUUID])

		if testCase != actual {
			t.Errorf("data not saved expected: %q, actual: %q", testCase, actual)
		}
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
			cleanEntries(t)
			entries.RLock()
			entries.m[testCase.UUID] = []byte(testCase.Value)
			entries.RUnlock()

			req := httptest.NewRequest("GET", fmt.Sprintf("http://example.com/%s", testCase.UUID), nil)
			w := httptest.NewRecorder()
			getHandler(w, req)

			resp := w.Result()
			body, _ := ioutil.ReadAll(resp.Body)

			actual := string(body)

			if string(body) != testCase.Value {
				t.Errorf("data not read expected: %q, actual: %q", testCase.Value, actual)
			}
		})
	}

}
