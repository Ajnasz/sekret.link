package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMuxHandle(t *testing.T) {
	t.Run("cleanPath", func(t *testing.T) {
		testCases := []struct {
			Pattern  string
			Expected string
		}{
			{
				Pattern:  "/foo/",
				Expected: "/foo",
			},
			{
				Pattern:  "/foo",
				Expected: "/foo",
			},
			{
				Pattern:  "/",
				Expected: "/",
			},
			{
				Pattern:  "",
				Expected: "/",
			},
			{
				Pattern:  "foo/bar/baz",
				Expected: "/foo/bar/baz",
			},
			{
				Pattern:  "foo/bar/baz/",
				Expected: "/foo/bar/baz",
			},
		}

		for _, testCase := range testCases {
			actual := cleanPath(testCase.Pattern)
			if actual != testCase.Expected {
				t.Errorf("Expected %q, got %q", testCase.Expected, actual)
			}
		}
	})

	t.Run("getPatternPartLen", func(t *testing.T) {
		testCases := []struct {
			Pattern  string
			Expected int
		}{
			{
				Pattern:  "/foo/bar/baz",
				Expected: 3,
			},
			{
				Pattern:  "foobarbaz",
				Expected: 0,
			},
		}

		for _, testCase := range testCases {
			actual := getPatternPartLen(testCase.Pattern)
			if actual != testCase.Expected {
				t.Errorf("Expected %d, got %d", testCase.Expected, actual)
			}
		}
	})

	t.Run("appendSorted", func(t *testing.T) {
		getNoopHandler := func() http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
		}
		handlers := []MuxEntry{}
		handlers = appendSorted(handlers, newMuxEntry("/foo", getNoopHandler()))
		handlers = appendSorted(handlers, newMuxEntry("/foo/bar/baz/qux", getNoopHandler()))
		handlers = appendSorted(handlers, newMuxEntry("/foo/bar", getNoopHandler()))
		handlers = appendSorted(handlers, newMuxEntry("/", getNoopHandler()))

		testPaths := []string{"/foo/bar/baz/qux", "/foo/bar", "/foo", "/"}
		if len(handlers) != len(testPaths) {
			t.Errorf("Not all items appended")
		}

		for index, value := range testPaths {
			handler := handlers[index]
			if handler.pattern != value {
				t.Errorf("Wrong order, expected %q to be %q", handler.pattern, testPaths[index])
			}
		}

	})

	t.Run("isPatternMatching", func(t *testing.T) {
		testCases := []struct {
			Pattern  string
			Str      string
			Matching bool
		}{
			{
				"/foo/bar",
				"foo/bar",
				true,
			},
			// {
			// 	"/foo/bar",
			// 	"/foo/bar",
			// 	true,
			// },
			// {
			// 	"/foo/bar",
			// 	"/foo",
			// 	false,
			// },
			// {
			// 	"/foo",
			// 	"/foo/bar",
			// 	false,
			// },

			// {
			// 	"/foo/:bar",
			// 	"/foo/something",
			// 	true,
			// },
			// {
			// 	"/foo/:bar",
			// 	"/foo/something/else",
			// 	false,
			// },
			// {
			// 	"/foo/:bar/baz",
			// 	"/foo/something/baz",
			// 	true,
			// },
		}

		for _, testCase := range testCases {
			isMatching := isPatternMatching(testCase.Pattern, testCase.Str)
			if isMatching != testCase.Matching {
				if testCase.Matching {
					t.Errorf("Wrong match, expected %q to match with %q", testCase.Pattern, testCase.Str)
				} else {
					t.Errorf("Wrong match, expected %q to NOT match with %q", testCase.Pattern, testCase.Str)
				}
			}
		}
	})

	t.Run("parsePatternMatch", func(t *testing.T) {
		mapsEqual := func(map1, map2 map[string]string) bool {
			if len(map1) != len(map2) {
				return false
			}

			for k, v := range map1 {
				v2, ok := map2[k]

				if !ok {
					return false
				}

				if v2 != v {
					return false
				}
			}

			return true
		}

		testCases := []struct {
			Pattern  string
			Path     string
			Expected map[string]string
		}{
			{
				Pattern: "/:foo/:bar/:baz",
				Path:    "/foo1/bar1/baz1",
				Expected: map[string]string{
					"foo": "foo1",
					"bar": "bar1",
					"baz": "baz1",
				},
			},
			// {
			// 	Pattern: "/lorem/:foo/:bar/:baz",
			// 	Path:    "/lorem/foo2/bar2/baz2",
			// 	Expected: map[string]string{
			// 		"foo": "foo2",
			// 		"bar": "bar2",
			// 		"baz": "baz2",
			// 	},
			// },
			// {
			// 	Pattern: "/lorem/:foo/ipsum/:bar/dolor/:baz/sit",
			// 	Path:    "/lorem/foo3/ipsum/bar3/dolor/baz3/sit",
			// 	Expected: map[string]string{
			// 		"foo": "foo2",
			// 		"bar": "bar2",
			// 		"baz": "baz2",
			// 	},
			// },
		}

		for _, testCase := range testCases {
			ret := parsePatternMatch(testCase.Pattern, testCase.Path)

			if !(mapsEqual(ret, testCase.Expected)) {
				t.Errorf("expected %+v to be equal with %+v", ret, testCase.Expected)
			}
		}
	})

	t.Run("Mux", func(t *testing.T) {
		t.Run("GetVars", func(t *testing.T) {
			mux := Mux{}

			stub := func() func(http.ResponseWriter, *http.Request) {
				return func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				}
			}()

			mux.Handle("/foo/:bar", http.HandlerFunc(stub))

			req := httptest.NewRequest(http.MethodGet, "http://example.com/foo/bar", nil)
			vars := mux.GetVars(req)
			bar, ok := vars["bar"]

			if !ok {
				t.Errorf("Expected to parse vars out from request")
			}

			if bar != "bar" {
				t.Errorf("Expected %q to be equal to %q", bar, "bar")
			}

			req = httptest.NewRequest(http.MethodGet, "http://example.com/foo/bar/baz", nil)
			vars = mux.GetVars(req)

			if vars != nil {
				t.Errorf("Expected %+v to be a nil", vars)
			}
		})

		t.Run("ServeHTTP", func(t *testing.T) {
			mux := Mux{}

			calls := []string{}
			stub := func() func(http.ResponseWriter, *http.Request) {
				return func(w http.ResponseWriter, r *http.Request) {
					calls = append(calls, r.URL.Path)
					w.WriteHeader(http.StatusOK)
				}
			}()

			mux.Handle("/foo/:bar", http.HandlerFunc(stub))

			req := httptest.NewRequest(http.MethodGet, "http://example.com/foo/bar", nil)
			mux.ServeHTTP(httptest.NewRecorder(), req)
			req2 := httptest.NewRequest(http.MethodGet, "http://example.com/foo/bar/baz", nil)
			mux.ServeHTTP(httptest.NewRecorder(), req2)

			if len(calls) != 1 {
				t.Errorf("Expected to cull stub 1 times, but was called %d", len(calls))
			}
		})
	})

	t.Run("MuxEntry", func(t *testing.T) {
		stub := func() func(http.ResponseWriter, *http.Request) {
			return func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}
		}()
		entry := newMuxEntry("/foo/bar", http.HandlerFunc(stub))
		entry.Method(http.MethodGet)

		if entry.method != http.MethodGet {
			t.Errorf("Expected method to be %q but got %q", http.MethodGet, entry.method)
		}
	})
}
