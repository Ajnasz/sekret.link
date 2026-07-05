package middlewares

import (
	"testing"
)

func TestRedactPath(t *testing.T) {
	testCases := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "root",
			path:     "/",
			expected: "/",
		},
		{
			name:     "empty",
			path:     "",
			expected: "",
		},
		{
			name:     "secret read path",
			path:     "/a34a0f8e-cdb9-4e3a-bd91-5f4023b1b210/0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
			expected: "/:redacted/:redacted",
		},
		{
			name:     "delete path",
			path:     "/a34a0f8e-cdb9-4e3a-bd91-5f4023b1b210/0123456789abcdef/deletekey",
			expected: "/:redacted/:redacted/:redacted",
		},
		{
			name:     "key generation path keeps the key literal",
			path:     "/key/a34a0f8e-cdb9-4e3a-bd91-5f4023b1b210/0123456789abcdef",
			expected: "/key/:redacted/:redacted",
		},
		{
			name:     "trailing slash near-miss",
			path:     "/a34a0f8e-cdb9-4e3a-bd91-5f4023b1b210/0123456789abcdef/",
			expected: "/:redacted/:redacted/",
		},
		{
			name:     "doubled slash near-miss",
			path:     "//a34a0f8e-cdb9-4e3a-bd91-5f4023b1b210//0123456789abcdef",
			expected: "//:redacted//:redacted",
		},
		{
			name:     "unknown literal is not trusted",
			path:     "/api/whatever",
			expected: "/:redacted/:redacted",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if actual := RedactPath(tc.path); actual != tc.expected {
				t.Errorf("RedactPath(%q) = %q, expected %q", tc.path, actual, tc.expected)
			}
		})
	}
}
