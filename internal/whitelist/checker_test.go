package whitelist

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-logr/logr"
)

func TestCheckSendsAPIKey(t *testing.T) {
	var gotKey, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotKey = r.Header.Get("X-API-Key")
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := NewChecker(logr.Discard(), "test-key")
	if res := c.Check(context.Background(), "user", "uuid", srv.URL, 5, 1); res != Allowed {
		t.Fatalf("expected Allowed, got %v", res)
	}
	if gotKey != "test-key" {
		t.Errorf("X-API-Key = %q, want %q", gotKey, "test-key")
	}
	if gotPath != "/api/whitelist" {
		t.Errorf("request path = %q, want %q", gotPath, "/api/whitelist")
	}
}
