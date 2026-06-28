package jenkins

import (
	"context"
	"net/http"
	"testing"
)

func TestListQueue(t *testing.T) {
	c, stop := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/queue/api/json" {
			t.Errorf("path=%s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"items":[{"id":42,"why":"Waiting for next available executor","stuck":false,"inQueueSince":1700000000000,"task":{"name":"build-it","url":"u"}}]}`))
	})
	defer stop()
	q, err := c.ListQueue(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(q) != 1 || q[0].ID != 42 || q[0].Task.Name != "build-it" {
		t.Errorf("got %+v", q)
	}
}
