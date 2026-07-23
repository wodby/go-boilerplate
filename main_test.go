package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandler(t *testing.T) {
	tests := []struct {
		name   string
		path   string
		status int
		body   string
	}{
		{name: "home", path: "/", status: http.StatusOK, body: "Hello from Wodby\n"},
		{name: "health", path: "/healthz", status: http.StatusOK, body: "ok\n"},
		{name: "not found", path: "/missing", status: http.StatusNotFound, body: "404 page not found\n"},
	}

	handler := newHandler()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, test.path, nil)
			response := httptest.NewRecorder()

			handler.ServeHTTP(response, request)

			result := response.Result()
			defer result.Body.Close()
			body, err := io.ReadAll(result.Body)
			if err != nil {
				t.Fatalf("read response: %v", err)
			}
			if result.StatusCode != test.status {
				t.Fatalf("status = %d, want %d", result.StatusCode, test.status)
			}
			if got := string(body); got != test.body {
				t.Fatalf("body = %q, want %q", got, test.body)
			}
			if got := result.Header.Get("Content-Type"); got != "text/plain; charset=utf-8" {
				t.Fatalf("content type = %q", got)
			}
		})
	}
}
