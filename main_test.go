package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestHandler(t *testing.T) {
	tests := []struct {
		name        string
		method      string
		path        string
		status      int
		body        string
		contentType string
	}{
		{
			name:        "home",
			method:      http.MethodGet,
			path:        "/",
			status:      http.StatusOK,
			body:        "Your Go app is running",
			contentType: "text/html; charset=utf-8",
		},
		{
			name:        "asset",
			method:      http.MethodGet,
			path:        "/assets/styles.css",
			status:      http.StatusOK,
			body:        ":root",
			contentType: "text/css; charset=utf-8",
		},
		{
			name:        "health",
			method:      http.MethodGet,
			path:        "/healthz",
			status:      http.StatusOK,
			body:        "ok\n",
			contentType: "text/plain; charset=utf-8",
		},
		{
			name:        "not found",
			method:      http.MethodGet,
			path:        "/missing",
			status:      http.StatusNotFound,
			body:        "404 page not found\n",
			contentType: "text/plain; charset=utf-8",
		},
		{
			name:        "method not allowed",
			method:      http.MethodPost,
			path:        "/",
			status:      http.StatusMethodNotAllowed,
			body:        "Method Not Allowed\n",
			contentType: "text/plain; charset=utf-8",
		},
	}

	handler := newHandler()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.method, test.path, nil)
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
			if got := string(body); !strings.Contains(got, test.body) {
				t.Fatalf("body = %q, want content %q", got, test.body)
			}
			if got := result.Header.Get("Content-Type"); got != test.contentType {
				t.Fatalf("content type = %q, want %q", got, test.contentType)
			}
		})
	}
}

func TestStatus(t *testing.T) {
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/status", nil)

	newHandler().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	var payload map[string]string
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload["status"] != "ok" || payload["server"] != "net/http" {
		t.Fatalf("payload = %#v", payload)
	}
}

func TestServerTimeouts(t *testing.T) {
	server := newServer(":8080", newHandler())

	if server.ReadHeaderTimeout != 5*time.Second {
		t.Fatalf("read header timeout = %s", server.ReadHeaderTimeout)
	}
	if server.IdleTimeout != 60*time.Second {
		t.Fatalf("idle timeout = %s", server.IdleTimeout)
	}
}
