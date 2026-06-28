package jenkins

import (
	"context"
	"net/http"
	"strings"
	"testing"
)

func TestListBuilds(t *testing.T) {
	c, stop := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/job/build-it/api/json" {
			t.Errorf("path=%s", r.URL.Path)
		}
		if !strings.Contains(r.URL.RawQuery, "tree=builds") {
			t.Errorf("tree missing: %s", r.URL.RawQuery)
		}
		_, _ = w.Write([]byte(`{"builds":[{"number":7,"result":"SUCCESS","building":false,"duration":1200,"timestamp":1700000000000,"url":"u","displayName":"#7"}]}`))
	})
	defer stop()
	bs, err := c.ListBuilds(context.Background(), "build-it", 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(bs) != 1 || bs[0].Number != 7 || bs[0].Result != "SUCCESS" {
		t.Errorf("got %+v", bs)
	}
}

func TestGetBuild(t *testing.T) {
	c, stop := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/job/build-it/7/api/json" {
			t.Errorf("path=%s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"number":7,"result":"FAILURE","building":false,"duration":3500,"url":"u"}`))
	})
	defer stop()
	b, err := c.GetBuild(context.Background(), "build-it", 7)
	if err != nil {
		t.Fatal(err)
	}
	if b.Result != "FAILURE" || b.Duration != 3500 {
		t.Errorf("got %+v", b)
	}
}
