package output

import (
	"bytes"
	"strings"
	"testing"
)

func TestDetect(t *testing.T) {
	tests := []struct {
		name string
		tty  bool
		over string
		want Format
	}{
		{"tty default", true, "", FormatTable},
		{"pipe default", false, "", FormatJSON},
		{"override md", true, "md", FormatMD},
		{"override json", true, "json", FormatJSON},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Detect(tt.tty, tt.over); got != tt.want {
				t.Errorf("got=%q want=%q", got, tt.want)
			}
		})
	}
}

func TestRender_JSON(t *testing.T) {
	var b bytes.Buffer
	if err := Render(&b, FormatJSON, map[string]any{"hello": "world"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(b.String(), `"hello": "world"`) {
		t.Errorf("got: %s", b.String())
	}
}

func TestRender_Table(t *testing.T) {
	var b bytes.Buffer
	rows := []map[string]any{{"name": "build-it", "color": "blue"}, {"name": "deploy", "color": "red"}}
	if err := Render(&b, FormatTable, rows); err != nil {
		t.Fatal(err)
	}
	out := b.String()
	for _, want := range []string{"NAME", "COLOR", "build-it", "blue", "deploy", "red"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
}

func TestRender_MD(t *testing.T) {
	var b bytes.Buffer
	rows := []map[string]any{{"name": "build-it", "color": "blue"}}
	if err := Render(&b, FormatMD, rows); err != nil {
		t.Fatal(err)
	}
	out := b.String()
	for _, want := range []string{"| name", "| color", "| build-it", "| blue", "|---"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
}
