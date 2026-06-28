package jenkins

import "testing"

func TestJobPath(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"simple", "build-it", "/job/build-it"},
		{"folder", "team/service", "/job/team/job/service"},
		{"deep", "org/team/service/main", "/job/org/job/team/job/service/job/main"},
		{"url-encode", "feature/PR-12 with space", "/job/feature/job/PR-12%20with%20space"},
		{"trim slashes", "/team/service/", "/job/team/job/service"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := JobPath(tt.in); got != tt.want {
				t.Errorf("got=%q want=%q", got, tt.want)
			}
		})
	}
}
