package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	dir := t.TempDir()
	credPath := filepath.Join(dir, "creds.yaml")
	body := `default-context: prod
contexts:
  prod:
    url: https://jenkins.prod.example.com
    username: alice
    token: SECRET_PROD
  staging:
    url: https://jenkins.staging.example.com
    username: alice
    token: SECRET_STAGING
    insecure: true
`
	if err := os.WriteFile(credPath, []byte(body), 0600); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		ctx     string
		want    Credentials
		wantErr bool
	}{
		{"default", "", Credentials{URL: "https://jenkins.prod.example.com", Username: "alice", Token: "SECRET_PROD"}, false},
		{"named", "staging", Credentials{URL: "https://jenkins.staging.example.com", Username: "alice", Token: "SECRET_STAGING", Insecure: true}, false},
		{"missing", "ghost", Credentials{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Load(credPath, tt.ctx)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err=%v wantErr=%v", err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("got=%+v want=%+v", got, tt.want)
			}
		})
	}
}

func TestLoad_RedactsTokenInError(t *testing.T) {
	dir := t.TempDir()
	credPath := filepath.Join(dir, "creds.yaml")
	_ = os.WriteFile(credPath, []byte("this is not yaml: ::: SECRET_TOKEN_XYZ"), 0600)
	_, err := Load(credPath, "")
	if err == nil {
		t.Fatal("expected error")
	}
	if got := err.Error(); contains(got, "SECRET_TOKEN_XYZ") {
		t.Errorf("token leaked into error: %s", got)
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
