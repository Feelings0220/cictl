package jenkins

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kagent-dev/ciq/internal/config"
)

func newServer(t *testing.T, handler http.HandlerFunc) (*Client, func()) {
	t.Helper()
	srv := httptest.NewServer(handler)
	c, err := New(config.Credentials{URL: srv.URL, Username: "u", Token: "t"})
	if err != nil {
		t.Fatal(err)
	}
	return c, srv.Close
}

func TestListJobs_TopLevel(t *testing.T) {
	c, stop := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/json" || !strings.Contains(r.URL.RawQuery, "tree=jobs") {
			t.Errorf("unexpected req: %s ? %s", r.URL.Path, r.URL.RawQuery)
		}
		_, _ = w.Write([]byte(`{"jobs":[{"name":"build-it","url":"http://x/job/build-it/","color":"blue"}]}`))
	})
	defer stop()
	jobs, err := c.ListJobs(context.Background(), "")
	if err != nil {
		t.Fatal(err)
	}
	if len(jobs) != 1 || jobs[0].Name != "build-it" || jobs[0].Color != "blue" {
		t.Errorf("got %+v", jobs)
	}
}

func TestListJobs_Folder(t *testing.T) {
	c, stop := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/job/team/job/service/api/json" {
			t.Errorf("path=%s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"jobs":[]}`))
	})
	defer stop()
	_, err := c.ListJobs(context.Background(), "team/service")
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetJob(t *testing.T) {
	c, stop := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/job/build-it/api/json" {
			t.Errorf("path=%s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"name":"build-it","fullName":"build-it","color":"red","lastBuild":{"number":7,"url":"u"}}`))
	})
	defer stop()
	j, err := c.GetJob(context.Background(), "build-it")
	if err != nil {
		t.Fatal(err)
	}
	if j.Color != "red" || j.LastBuild == nil || j.LastBuild.Number != 7 {
		t.Errorf("got %+v", j)
	}
}

func TestGetJobConfig(t *testing.T) {
	c, stop := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/job/build-it/config.xml" {
			t.Errorf("path=%s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/xml")
		_, _ = w.Write([]byte(`<project><builders/></project>`))
	})
	defer stop()
	b, err := c.GetJobConfig(context.Background(), "build-it")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "<project>") {
		t.Errorf("got %s", string(b))
	}
}
