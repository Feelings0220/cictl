package jenkins

import (
	"context"
	"net/http"
	"strings"
	"testing"
)

func TestGetConsole(t *testing.T) {
	c, stop := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/job/build-it/7/consoleText" {
			t.Errorf("path=%s", r.URL.Path)
		}
		_, _ = w.Write([]byte("hello\nworld\n"))
	})
	defer stop()
	b, err := c.GetConsole(context.Background(), "build-it", 7)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != "hello\nworld\n" {
		t.Errorf("got %q", string(b))
	}
}

func TestLastNLines(t *testing.T) {
	in := []byte("a\nb\nc\nd\ne\n")
	tests := []struct {
		n    int
		want string
	}{
		{1, "e\n"},
		{3, "c\nd\ne\n"},
		{10, "a\nb\nc\nd\ne\n"},
		{0, ""},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := string(LastNLines(in, tt.n))
			if got != tt.want {
				t.Errorf("n=%d got=%q want=%q", tt.n, got, tt.want)
			}
		})
	}
	// no trailing newline
	if got := string(LastNLines([]byte("x\ny\nz"), 2)); !strings.HasSuffix(got, "z") {
		t.Errorf("got %q", got)
	}
}
