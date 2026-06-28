package jenkins

import (
	"net/url"
	"strings"
)

// JobPath converts "team/service/main" to "/job/team/job/service/job/main"
// with each segment URL-path-escaped.
func JobPath(name string) string {
	name = strings.Trim(name, "/")
	if name == "" {
		return ""
	}
	parts := strings.Split(name, "/")
	var b strings.Builder
	for _, p := range parts {
		b.WriteString("/job/")
		b.WriteString(url.PathEscape(p))
	}
	return b.String()
}
