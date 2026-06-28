package jenkins

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kagent-dev/ciq/internal/config"
)

func TestClient_GET(t *testing.T) {
	type sample struct {
		Hello string `json:"hello"`
	}

	tests := []struct {
		name       string
		status     int
		body       string
		path       string
		wantHello  string
		wantErrSub string
	}{
		{"ok", 200, `{"hello":"world"}`, "/api/json", "world", ""},
		{"not found", 404, `not found`, "/missing", "", "404"},
		{"server error", 500, `boom`, "/x", "", "500"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				want := "Basic " + base64.StdEncoding.EncodeToString([]byte("alice:TOKEN"))
				if got := r.Header.Get("Authorization"); got != want {
					t.Errorf("auth header got=%q want=%q", got, want)
				}
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer srv.Close()

			c, err := New(config.Credentials{URL: srv.URL, Username: "alice", Token: "TOKEN"})
			if err != nil {
				t.Fatal(err)
			}
			var dst sample
			_, err = c.GET(context.Background(), tt.path, &dst)
			if tt.wantErrSub == "" {
				if err != nil {
					t.Fatalf("unexpected err: %v", err)
				}
				if dst.Hello != tt.wantHello {
					t.Errorf("hello=%q want=%q", dst.Hello, tt.wantHello)
				}
				return
			}
			if err == nil {
				t.Fatal("want error")
			}
			if !strings.Contains(err.Error(), tt.wantErrSub) {
				t.Errorf("err=%q want substr %q", err.Error(), tt.wantErrSub)
			}
			if strings.Contains(err.Error(), "TOKEN") {
				t.Errorf("token leaked in error: %s", err.Error())
			}
		})
	}
}
