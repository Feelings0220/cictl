package jenkins

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Feelings0220/cictl/internal/config"
)

func TestWhoAmI(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/me/api/json" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"id":"alice","fullName":"Alice Example","authenticated":true}`))
	}))
	defer srv.Close()

	c, _ := New(config.Credentials{URL: srv.URL, Username: "alice", Token: "t"})
	me, err := c.WhoAmI(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	want := Me{ID: "alice", FullName: "Alice Example", Authenticated: true}
	if me != want {
		t.Errorf("got=%+v want=%+v", me, want)
	}
}
