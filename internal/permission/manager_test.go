package permission

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-logr/logr"
)

func TestFetchSendsAPIKey(t *testing.T) {
	var gotKey, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotKey = r.Header.Get("X-API-Key")
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"success":true,"users":[{"username":"user","permission_level":4}]}`))
	}))
	defer srv.Close()

	m := NewManager(logr.Discard(), srv.URL, "test-key", 300, []string{"send"})
	if level := m.GetPermissionLevel(context.Background(), "user"); level != 4 {
		t.Fatalf("permission level = %d, want 4", level)
	}
	if gotKey != "test-key" {
		t.Errorf("X-API-Key = %q, want %q", gotKey, "test-key")
	}
	if gotPath != "/api/mcdr/permission" {
		t.Errorf("request path = %q, want %q", gotPath, "/api/mcdr/permission")
	}
}
