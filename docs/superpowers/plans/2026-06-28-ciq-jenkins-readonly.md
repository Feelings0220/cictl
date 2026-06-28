# ciq Jenkins Read-Only CLI + Skill Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a standalone open-source CLI (`ciq`) plus a portable skill markdown that lets any AI agent (kagent, Claude Code, Cursor, …) safely query Jenkins read-only. Make it ready to vendor into kagent via a new `helm/agents/jenkins-triage/` chart submitted as PR. Separately deliver a Jenkins Shared Library guide for `kagentAnalyze()` (doc only, no repo).

**Architecture:** A small Go cobra binary that wraps the Jenkins REST API. Three layers: `internal/config` (credentials.yaml + context switching), `internal/jenkins` (HTTP client + endpoint methods, fully mocked in tests via `httptest`), `internal/output` (json/table/markdown formatters). Subcommands live in `cmd/ciq/`. Read-only is enforced **at compile time** — write/mutate subcommands simply do not exist in this build. A single portable `skills/jenkins.md` is shipped beside the binary so any agent can load it.

**Tech Stack:** Go 1.26+, cobra (CLI), gopkg.in/yaml.v3 (config), stdlib `net/http` (no Jenkins SDK), stdlib `testing` + `httptest` (no external test deps), Apache 2.0 license.

## Global Constraints

- **Go version:** 1.26 minimum (matches kagent `go/go.mod` `go 1.26.3`).
- **Read-only only:** This build MUST NOT compile any subcommand that triggers builds, aborts, replays, or modifies Jenkins state. Mutation subcommands are deferred to a future `--enable-mutations` build tag and are not part of this plan.
- **No JVM dependency:** Never shell out to `jenkins-cli.jar`. All operations are direct Jenkins REST.
- **License:** Apache 2.0, identical SPDX header style to kagent (no per-file copyright headers — kagent does not use them).
- **Error wrapping:** Every returned error wraps the cause with `fmt.Errorf("...: %w", err)`. Matches kagent `CLAUDE.md` convention.
- **Tests:** Table-driven Go tests. Every Jenkins endpoint method has a `httptest.NewServer` test. No live Jenkins calls in CI.
- **Commits:** Conventional Commits format (`feat:`, `fix:`, `docs:`, `test:`, `chore:`, `refactor:`).
- **Output:** Every read command supports `--format json|table|md`; default `table` for humans, `json` when stdout is not a TTY (LLM-friendly).
- **Auth:** `~/.config/ciq/credentials.yaml` with `--context <name>` flag; never log token; never include token in error messages.
- **No mutating HTTP verbs:** The client struct MUST only expose `GET`. There is no `POST`/`PUT`/`DELETE` helper. Compile-time guarantee.

---

## File Structure

```
E:\claudecode\mykagent\ciq\                     # new standalone OSS project
├── go.mod                                       # module github.com/kagent-dev/ciq (placeholder until owner decides)
├── go.sum
├── Makefile                                     # build, test, lint, vendor-to-kagent
├── README.md                                    # OSS-facing readme
├── LICENSE                                      # Apache 2.0
├── CONTRIBUTING.md                              # PR rules, conventional commits, sign-off
├── .gitignore
├── .github/workflows/ci.yaml                    # go build + test + golangci-lint
├── cmd/ciq/
│   ├── main.go                                  # entrypoint, calls cmd.Execute()
│   ├── root.go                                  # root cobra command, --context, --format flags
│   ├── jenkins.go                               # `ciq jenkins` parent command
│   ├── jenkins_whoami.go                        # `ciq jenkins whoami`
│   ├── jenkins_job.go                           # `ciq jenkins job list/get/config`
│   ├── jenkins_build.go                         # `ciq jenkins build list/get`
│   ├── jenkins_console.go                       # `ciq jenkins console <job> <num> [--tail N | --full]`
│   ├── jenkins_cloud.go                         # `ciq jenkins cloud list/get`
│   └── jenkins_queue.go                         # `ciq jenkins queue list`
├── internal/config/
│   ├── config.go                                # credentials loader, context resolver
│   └── config_test.go
├── internal/jenkins/
│   ├── client.go                                # HTTP client struct (GET only)
│   ├── client_test.go
│   ├── path.go                                  # folder-path → URL helper (jobs nested in folders)
│   ├── path_test.go
│   ├── jobs.go                                  # ListJobs, GetJob, GetJobConfig
│   ├── jobs_test.go
│   ├── builds.go                                # ListBuilds, GetBuild
│   ├── builds_test.go
│   ├── console.go                               # GetConsoleText (with offset/tail)
│   ├── console_test.go
│   ├── clouds.go                                # ListClouds, GetCloudConfig (kubernetes-plugin)
│   ├── clouds_test.go
│   ├── queue.go                                 # ListQueue
│   ├── queue_test.go
│   ├── whoami.go                                # WhoAmI
│   └── whoami_test.go
├── internal/output/
│   ├── formatter.go                             # JSON / table / md renderers
│   └── formatter_test.go
├── skills/
│   └── jenkins.md                               # agent-facing skill, kagent-loadable via gitRef
├── examples/
│   ├── credentials.yaml.example
│   └── jenkins-triage-agent.yaml                # sample kagent Agent CRD that uses this skill
├── docs/
│   ├── architecture.md
│   ├── kagent-pr-checklist.md                   # how to vendor into kagent helm/agents/
│   └── superpowers/plans/2026-06-28-ciq-jenkins-readonly.md   # this file
└── kagent-pr/                                    # staging area for the kagent PR (Helm chart)
    └── helm-agents-jenkins-triage/
        ├── Chart-template.yaml
        ├── values.yaml
        └── templates/
            ├── _helpers.tpl
            └── agent.yaml

# delivered separately, NOT in any repo:
E:\claudecode\mykagent\jenkins-shared-library-guide.md
```

**Decomposition rationale:**
- One file per Jenkins resource (jobs, builds, console, clouds, queue) so a reviewer can accept/reject independently.
- `client.go` has zero endpoint knowledge — only `GET()`, auth, error mapping. Each endpoint file talks to the client.
- `cmd/ciq/jenkins_*.go` files have zero HTTP knowledge — only flag parsing + delegating to `internal/jenkins`.
- The `kagent-pr/` directory is staging only; final destination in the kagent repo is `helm/agents/jenkins-triage/`.

---

## Task 1: Project Scaffolding

**Files:**
- Create: `E:\claudecode\mykagent\ciq\go.mod`
- Create: `E:\claudecode\mykagent\ciq\Makefile`
- Create: `E:\claudecode\mykagent\ciq\README.md`
- Create: `E:\claudecode\mykagent\ciq\LICENSE`
- Create: `E:\claudecode\mykagent\ciq\.gitignore`
- Create: `E:\claudecode\mykagent\ciq\.github\workflows\ci.yaml`
- Create: `E:\claudecode\mykagent\ciq\CONTRIBUTING.md`

**Interfaces:**
- Consumes: nothing
- Produces: a buildable empty Go module at `github.com/kagent-dev/ciq` (placeholder path).

- [ ] **Step 1: Create `go.mod`**

```
module github.com/kagent-dev/ciq

go 1.26
```

- [ ] **Step 2: Create `LICENSE`** — copy Apache 2.0 text verbatim from https://www.apache.org/licenses/LICENSE-2.0.txt (the kagent repo uses the standard form).

- [ ] **Step 3: Create `.gitignore`**

```
/dist/
/ciq
/ciq.exe
*.test
*.out
coverage.txt
.idea/
.vscode/
```

- [ ] **Step 4: Create `Makefile`**

```makefile
BINARY := ciq
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-s -w -X main.version=$(VERSION)"

.PHONY: build test lint clean vendor-to-kagent

build:
	go build $(LDFLAGS) -o $(BINARY) ./cmd/ciq

test:
	go test -race -count=1 ./...

lint:
	golangci-lint run ./...

clean:
	rm -f $(BINARY) $(BINARY).exe coverage.txt

# Stage the helm chart + skill into a kagent worktree at $(KAGENT_DIR).
# Used to produce the kagent PR. See docs/kagent-pr-checklist.md.
vendor-to-kagent:
	@test -n "$(KAGENT_DIR)" || (echo "set KAGENT_DIR=path/to/kagent"; exit 1)
	mkdir -p $(KAGENT_DIR)/helm/agents/jenkins-triage/templates
	cp -r kagent-pr/helm-agents-jenkins-triage/* $(KAGENT_DIR)/helm/agents/jenkins-triage/
```

- [ ] **Step 5: Create `.github\workflows\ci.yaml`**

```yaml
name: CI
on:
  push: { branches: [main] }
  pull_request:
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: '1.26' }
      - run: go build ./...
      - run: go test -race -count=1 ./...
      - uses: golangci/golangci-lint-action@v6
        with: { version: latest }
```

- [ ] **Step 6: Create `README.md`** with placeholder body (full content written in Task 13):

```markdown
# ciq — AI-agent-friendly CI inspection CLI

Read-only CLI for querying CI/CD systems (Jenkins first; GitLab CI / GitHub Actions planned).
Designed to be safe for AI agents to invoke: structured JSON output, compile-time read-only guarantee,
auth tokens never leak into command output or LLM context.

Status: alpha, Jenkins read-only commands.

(Full README in Task 13.)
```

- [ ] **Step 7: Create `CONTRIBUTING.md`**

```markdown
# Contributing

- Conventional Commits required: `feat:`, `fix:`, `docs:`, `test:`, `chore:`, `refactor:`.
- Sign off commits with `git commit -s`.
- Every new Jenkins endpoint method needs an `httptest`-backed table-driven test.
- No mutating HTTP verbs in `internal/jenkins/client.go` without an RFC-style design issue first.
```

- [ ] **Step 8: Verify build works**

Run: `cd E:\claudecode\mykagent\ciq && go build ./...`
Expected: succeeds with no source files (Go is happy with an empty module).

- [ ] **Step 9: Init git + initial commit**

```bash
git init
git add .
git commit -s -m "chore: initial project scaffolding"
```

---

## Task 2: Config Loader

**Files:**
- Create: `E:\claudecode\mykagent\ciq\internal\config\config.go`
- Create: `E:\claudecode\mykagent\ciq\internal\config\config_test.go`
- Create: `E:\claudecode\mykagent\ciq\examples\credentials.yaml.example`

**Interfaces:**
- Consumes: nothing
- Produces:
  ```go
  type Credentials struct {
      URL      string // Jenkins base URL, e.g. https://jenkins.example.com
      Username string
      Token    string // API token, never logged
      Insecure bool   // skip TLS verify (lab use)
  }

  func Load(path string, context string) (Credentials, error)
  // path: explicit credentials file path, "" → default $XDG_CONFIG_HOME/ciq/credentials.yaml
  // context: named context to pick; "" → use file's `default-context` key
  ```

- [ ] **Step 1: Write the failing test** — `internal/config/config_test.go`

```go
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
		{"named",   "staging", Credentials{URL: "https://jenkins.staging.example.com", Username: "alice", Token: "SECRET_STAGING", Insecure: true}, false},
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
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/config -run TestLoad -v`
Expected: FAIL — `Load` not defined.

- [ ] **Step 3: Add yaml dep**

```bash
go get gopkg.in/yaml.v3
go mod tidy
```

- [ ] **Step 4: Write `internal/config/config.go`**

```go
// Package config loads ciq credentials from a YAML file.
package config

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Credentials struct {
	URL      string
	Username string
	Token    string
	Insecure bool
}

type fileContext struct {
	URL      string `yaml:"url"`
	Username string `yaml:"username"`
	Token    string `yaml:"token"`
	Insecure bool   `yaml:"insecure"`
}

type file struct {
	DefaultContext string                 `yaml:"default-context"`
	Contexts       map[string]fileContext `yaml:"contexts"`
}

func Load(path, context string) (Credentials, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Credentials{}, fmt.Errorf("read credentials: %w", err)
	}
	var f file
	if err := yaml.Unmarshal(raw, &f); err != nil {
		// Never echo file content (may contain token) in error.
		return Credentials{}, errors.New("parse credentials yaml: invalid syntax")
	}
	if context == "" {
		context = f.DefaultContext
	}
	c, ok := f.Contexts[context]
	if !ok {
		return Credentials{}, fmt.Errorf("context %q not found in credentials", context)
	}
	return Credentials{URL: c.URL, Username: c.Username, Token: c.Token, Insecure: c.Insecure}, nil
}
```

- [ ] **Step 5: Run test to verify it passes**

Run: `go test ./internal/config -v`
Expected: PASS (both tests).

- [ ] **Step 6: Create `examples/credentials.yaml.example`**

```yaml
default-context: prod
contexts:
  prod:
    url: https://jenkins.prod.example.com
    username: alice
    # Generate via Jenkins → People → alice → Configure → API Token.
    token: REPLACE_ME
  staging:
    url: https://jenkins.staging.example.com
    username: alice
    token: REPLACE_ME
    insecure: true  # lab clusters with self-signed certs only
```

- [ ] **Step 7: Commit**

```bash
git add internal/config examples/credentials.yaml.example go.mod go.sum
git commit -s -m "feat(config): credentials loader with named contexts"
```

---

## Task 3: HTTP Client Base

**Files:**
- Create: `E:\claudecode\mykagent\ciq\internal\jenkins\client.go`
- Create: `E:\claudecode\mykagent\ciq\internal\jenkins\client_test.go`

**Interfaces:**
- Consumes: `config.Credentials` from Task 2.
- Produces:
  ```go
  type Client struct { /* unexported */ }
  func New(c config.Credentials) (*Client, error)
  // GET performs an authenticated GET to a Jenkins path (path begins with "/").
  // If into != nil, response is decoded as JSON into into.
  // Always returns raw bytes for callers that want non-JSON (config.xml, consoleText).
  func (c *Client) GET(ctx context.Context, path string, into any) ([]byte, error)
  ```

- [ ] **Step 1: Write the failing test** — covers (a) basic auth header, (b) 200 → JSON decode, (c) 4xx → typed error, (d) token never appears in error.

```go
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
	type sample struct{ Hello string `json:"hello"` }

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
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/jenkins -run TestClient_GET -v`
Expected: FAIL — `New` / `Client` undefined.

- [ ] **Step 3: Write `internal/jenkins/client.go`**

```go
// Package jenkins is a minimal read-only Jenkins REST client.
package jenkins

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/kagent-dev/ciq/internal/config"
)

type Client struct {
	base  *url.URL
	user  string
	token string
	http  *http.Client
}

func New(c config.Credentials) (*Client, error) {
	if c.URL == "" {
		return nil, fmt.Errorf("jenkins url is empty")
	}
	u, err := url.Parse(strings.TrimRight(c.URL, "/"))
	if err != nil {
		return nil, fmt.Errorf("parse jenkins url: %w", err)
	}
	tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: c.Insecure}} //nolint:gosec // opt-in via config
	return &Client{
		base:  u,
		user:  c.Username,
		token: c.Token,
		http:  &http.Client{Transport: tr, Timeout: 30 * time.Second},
	}, nil
}

// GET performs an authenticated GET. path begins with "/".
// If into is non-nil and response is JSON, it is unmarshalled into it.
// Raw response body is always returned for callers that want bytes.
func (c *Client) GET(ctx context.Context, path string, into any) ([]byte, error) {
	if !strings.HasPrefix(path, "/") {
		return nil, fmt.Errorf("path must begin with /")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.base.String()+path, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.SetBasicAuth(c.user, c.token)
	req.Header.Set("Accept", "application/json,application/xml,text/plain;q=0.9,*/*;q=0.5")
	resp, err := c.http.Do(req)
	if err != nil {
		// Strip credentials from any url errors.
		return nil, fmt.Errorf("jenkins request failed: %w", scrubURLError(err, c.token))
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}
	if resp.StatusCode >= 400 {
		return body, fmt.Errorf("jenkins returned %d for %s", resp.StatusCode, path)
	}
	if into != nil {
		if err := json.Unmarshal(body, into); err != nil {
			return body, fmt.Errorf("decode json from %s: %w", path, err)
		}
	}
	return body, nil
}

// scrubURLError removes the token from net/url errors that embed the request URL.
func scrubURLError(err error, token string) error {
	if token == "" {
		return err
	}
	return fmt.Errorf("%s", strings.ReplaceAll(err.Error(), token, "***"))
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/jenkins -run TestClient_GET -v`
Expected: PASS (3 sub-tests).

- [ ] **Step 5: Commit**

```bash
git add internal/jenkins/client.go internal/jenkins/client_test.go
git commit -s -m "feat(jenkins): read-only http client with auth and error scrubbing"
```

---

## Task 4: Folder-Path Helper

**Files:**
- Create: `E:\claudecode\mykagent\ciq\internal\jenkins\path.go`
- Create: `E:\claudecode\mykagent\ciq\internal\jenkins\path_test.go`

**Interfaces:**
- Consumes: nothing.
- Produces: `func JobPath(name string) string` — turns a slash-separated job name (`team/service/main`) into Jenkins' nested `/job/<a>/job/<b>/job/<c>` URL form.

- [ ] **Step 1: Failing test**

```go
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
```

- [ ] **Step 2: Run, verify FAIL**

Run: `go test ./internal/jenkins -run TestJobPath -v`

- [ ] **Step 3: Implement** — `internal/jenkins/path.go`

```go
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
```

- [ ] **Step 4: Run, verify PASS**

Run: `go test ./internal/jenkins -run TestJobPath -v`

- [ ] **Step 5: Commit**

```bash
git add internal/jenkins/path.go internal/jenkins/path_test.go
git commit -s -m "feat(jenkins): nested folder path helper"
```

---

## Task 5: Output Formatter

**Files:**
- Create: `E:\claudecode\mykagent\ciq\internal\output\formatter.go`
- Create: `E:\claudecode\mykagent\ciq\internal\output\formatter_test.go`

**Interfaces:**
- Consumes: nothing.
- Produces:
  ```go
  type Format string
  const (FormatJSON Format = "json"; FormatTable Format = "table"; FormatMD Format = "md")
  func Detect(stdoutIsTTY bool, override string) Format
  // Render writes value to w. For table/md the value must be []map[string]any or struct slice;
  // for json any value is fine.
  func Render(w io.Writer, f Format, value any) error
  ```

- [ ] **Step 1: Failing test**

```go
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
```

- [ ] **Step 2: Run, FAIL**

Run: `go test ./internal/output -v`

- [ ] **Step 3: Implement** — `internal/output/formatter.go`

```go
// Package output renders structured values for humans and agents.
package output

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"text/tabwriter"
)

type Format string

const (
	FormatJSON  Format = "json"
	FormatTable Format = "table"
	FormatMD    Format = "md"
)

func Detect(stdoutIsTTY bool, override string) Format {
	if override != "" {
		return Format(override)
	}
	if stdoutIsTTY {
		return FormatTable
	}
	return FormatJSON
}

func Render(w io.Writer, f Format, value any) error {
	switch f {
	case FormatJSON:
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(value)
	case FormatTable:
		return renderTable(w, value)
	case FormatMD:
		return renderMD(w, value)
	}
	return fmt.Errorf("unsupported format: %s", f)
}

func rowsOf(value any) ([]map[string]any, error) {
	rows, ok := value.([]map[string]any)
	if !ok {
		return nil, fmt.Errorf("value must be []map[string]any for table/md")
	}
	return rows, nil
}

func columns(rows []map[string]any) []string {
	seen := map[string]struct{}{}
	for _, r := range rows {
		for k := range r {
			seen[k] = struct{}{}
		}
	}
	cols := make([]string, 0, len(seen))
	for k := range seen {
		cols = append(cols, k)
	}
	sort.Strings(cols)
	return cols
}

func renderTable(w io.Writer, value any) error {
	rows, err := rowsOf(value)
	if err != nil {
		return err
	}
	cols := columns(rows)
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	headers := make([]string, len(cols))
	for i, c := range cols {
		headers[i] = strings.ToUpper(c)
	}
	fmt.Fprintln(tw, strings.Join(headers, "\t"))
	for _, r := range rows {
		cells := make([]string, len(cols))
		for i, c := range cols {
			cells[i] = fmt.Sprintf("%v", r[c])
		}
		fmt.Fprintln(tw, strings.Join(cells, "\t"))
	}
	return tw.Flush()
}

func renderMD(w io.Writer, value any) error {
	rows, err := rowsOf(value)
	if err != nil {
		return err
	}
	cols := columns(rows)
	fmt.Fprintln(w, "| "+strings.Join(cols, " | ")+" |")
	sep := make([]string, len(cols))
	for i := range cols {
		sep[i] = "---"
	}
	fmt.Fprintln(w, "|"+strings.Join(sep, "|")+"|")
	for _, r := range rows {
		cells := make([]string, len(cols))
		for i, c := range cols {
			cells[i] = fmt.Sprintf("%v", r[c])
		}
		fmt.Fprintln(w, "| "+strings.Join(cells, " | ")+" |")
	}
	return nil
}
```

- [ ] **Step 4: Run, PASS**

Run: `go test ./internal/output -v`

- [ ] **Step 5: Commit**

```bash
git add internal/output
git commit -s -m "feat(output): json/table/md formatters"
```

---

## Task 6: Cobra Root + `ciq jenkins whoami`

This task wires up `cmd/ciq` end-to-end and adds the simplest Jenkins endpoint (`/me/api/json`) as a smoke test that proves auth + client + formatter together.

**Files:**
- Create: `E:\claudecode\mykagent\ciq\cmd\ciq\main.go`
- Create: `E:\claudecode\mykagent\ciq\cmd\ciq\root.go`
- Create: `E:\claudecode\mykagent\ciq\cmd\ciq\jenkins.go`
- Create: `E:\claudecode\mykagent\ciq\cmd\ciq\jenkins_whoami.go`
- Create: `E:\claudecode\mykagent\ciq\internal\jenkins\whoami.go`
- Create: `E:\claudecode\mykagent\ciq\internal\jenkins\whoami_test.go`

**Interfaces:**
- Consumes: `jenkins.Client.GET` (Task 3), `output.Render` (Task 5), `config.Load` (Task 2).
- Produces:
  ```go
  // internal/jenkins
  type Me struct { ID, FullName string; Authenticated bool }
  func (c *Client) WhoAmI(ctx context.Context) (Me, error)
  ```
  CLI: `ciq jenkins whoami` prints the authenticated user.

- [ ] **Step 1: Add cobra dep**

```bash
go get github.com/spf13/cobra
go mod tidy
```

- [ ] **Step 2: Failing test for `WhoAmI`**

```go
package jenkins

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kagent-dev/ciq/internal/config"
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
```

- [ ] **Step 3: Run, FAIL**

Run: `go test ./internal/jenkins -run TestWhoAmI -v`

- [ ] **Step 4: Implement `WhoAmI`** — `internal/jenkins/whoami.go`

```go
package jenkins

import "context"

type Me struct {
	ID            string `json:"id"`
	FullName      string `json:"fullName"`
	Authenticated bool   `json:"authenticated"`
}

func (c *Client) WhoAmI(ctx context.Context) (Me, error) {
	var m Me
	if _, err := c.GET(ctx, "/me/api/json", &m); err != nil {
		return Me{}, err
	}
	return m, nil
}
```

- [ ] **Step 5: Run, PASS**

Run: `go test ./internal/jenkins -run TestWhoAmI -v`

- [ ] **Step 6: Write `cmd/ciq/main.go`**

```go
package main

import (
	"fmt"
	"os"
)

var version = "dev"

func main() {
	if err := newRoot().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
```

- [ ] **Step 7: Write `cmd/ciq/root.go`**

```go
package main

import (
	"os"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

type rootFlags struct {
	credsFile string
	context   string
	format    string
}

var rf rootFlags

func newRoot() *cobra.Command {
	root := &cobra.Command{
		Use:           "ciq",
		Short:         "AI-agent-friendly CI inspection CLI (read-only)",
		SilenceUsage:  true,
		SilenceErrors: false,
		Version:       version,
	}
	defaultCreds, _ := os.UserConfigDir()
	if defaultCreds != "" {
		defaultCreds += string(os.PathSeparator) + "ciq" + string(os.PathSeparator) + "credentials.yaml"
	}
	root.PersistentFlags().StringVar(&rf.credsFile, "credentials", defaultCreds, "path to credentials.yaml")
	root.PersistentFlags().StringVar(&rf.context, "context", "", "credentials context (defaults to default-context)")
	root.PersistentFlags().StringVar(&rf.format, "format", "", "output format: json|table|md (default: table on tty, json otherwise)")
	root.AddCommand(newJenkinsCmd())
	return root
}

func stdoutIsTTY() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}
```

- [ ] **Step 8: Add `golang.org/x/term` dep**

```bash
go get golang.org/x/term
go mod tidy
```

- [ ] **Step 9: Write `cmd/ciq/jenkins.go`**

```go
package main

import (
	"context"
	"fmt"

	"github.com/kagent-dev/ciq/internal/config"
	"github.com/kagent-dev/ciq/internal/jenkins"
	"github.com/spf13/cobra"
)

func newJenkinsCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "jenkins",
		Short: "Query Jenkins (read-only)",
	}
	c.AddCommand(newJenkinsWhoAmICmd())
	return c
}

// newJenkinsClient resolves credentials and constructs a Jenkins client.
// Shared by all jenkins_*.go subcommands.
func newJenkinsClient(_ *cobra.Command) (*jenkins.Client, context.Context, error) {
	if rf.credsFile == "" {
		return nil, nil, fmt.Errorf("no credentials file (set --credentials or place credentials.yaml under your user config dir)")
	}
	creds, err := config.Load(rf.credsFile, rf.context)
	if err != nil {
		return nil, nil, err
	}
	cl, err := jenkins.New(creds)
	if err != nil {
		return nil, nil, err
	}
	return cl, context.Background(), nil
}
```

- [ ] **Step 10: Write `cmd/ciq/jenkins_whoami.go`**

```go
package main

import (
	"os"

	"github.com/kagent-dev/ciq/internal/output"
	"github.com/spf13/cobra"
)

func newJenkinsWhoAmICmd() *cobra.Command {
	return &cobra.Command{
		Use:   "whoami",
		Short: "Show the authenticated user",
		RunE: func(cmd *cobra.Command, args []string) error {
			cl, ctx, err := newJenkinsClient(cmd)
			if err != nil {
				return err
			}
			me, err := cl.WhoAmI(ctx)
			if err != nil {
				return err
			}
			rows := []map[string]any{{"id": me.ID, "fullName": me.FullName, "authenticated": me.Authenticated}}
			return output.Render(os.Stdout, output.Detect(stdoutIsTTY(), rf.format), rows)
		},
	}
}
```

- [ ] **Step 11: Build, smoke-test help works**

Run: `go build ./cmd/ciq && ./ciq --help && ./ciq jenkins --help && ./ciq jenkins whoami --help`
Expected: All three help screens render. (No live Jenkins call.)

- [ ] **Step 12: Commit**

```bash
git add cmd internal/jenkins/whoami.go internal/jenkins/whoami_test.go go.mod go.sum
git commit -s -m "feat(cli): cobra root and jenkins whoami subcommand"
```

---

## Task 7: Jobs Subcommands (`list`, `get`, `config`)

**Files:**
- Create: `E:\claudecode\mykagent\ciq\internal\jenkins\jobs.go`
- Create: `E:\claudecode\mykagent\ciq\internal\jenkins\jobs_test.go`
- Create: `E:\claudecode\mykagent\ciq\cmd\ciq\jenkins_job.go`
- Modify: `E:\claudecode\mykagent\ciq\cmd\ciq\jenkins.go` — register `job` subcommand.

**Interfaces:**
- Consumes: `Client.GET`, `JobPath`.
- Produces:
  ```go
  type JobSummary struct { Name, URL, Color string }
  type Job struct {
      Name, FullName, URL, Description string
      Buildable     bool
      Color         string
      InQueue       bool
      HealthReport  []HealthReport `json:"healthReport"`
      Builds        []BuildRef
      LastBuild     *BuildRef `json:"lastBuild"`
      LastCompletedBuild *BuildRef `json:"lastCompletedBuild"`
      LastFailedBuild    *BuildRef `json:"lastFailedBuild"`
  }
  type HealthReport struct { Description string; Score int }
  type BuildRef struct { Number int; URL string }

  func (c *Client) ListJobs(ctx context.Context, folder string) ([]JobSummary, error)
  // folder is "" for top-level, or "team/service" for nested folders.
  func (c *Client) GetJob(ctx context.Context, name string) (Job, error)
  // Returns raw config.xml bytes.
  func (c *Client) GetJobConfig(ctx context.Context, name string) ([]byte, error)
  ```

- [ ] **Step 1: Failing tests** — `internal/jenkins/jobs_test.go`

```go
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
```

- [ ] **Step 2: Run, FAIL**

Run: `go test ./internal/jenkins -run TestListJobs -run TestGetJob -v`

- [ ] **Step 3: Implement** — `internal/jenkins/jobs.go`

```go
package jenkins

import (
	"context"
	"fmt"
)

type JobSummary struct {
	Name  string `json:"name"`
	URL   string `json:"url"`
	Color string `json:"color"`
}

type HealthReport struct {
	Description string `json:"description"`
	Score       int    `json:"score"`
}

type BuildRef struct {
	Number int    `json:"number"`
	URL    string `json:"url"`
}

type Job struct {
	Name               string         `json:"name"`
	FullName           string         `json:"fullName"`
	URL                string         `json:"url"`
	Description        string         `json:"description"`
	Buildable          bool           `json:"buildable"`
	Color              string         `json:"color"`
	InQueue            bool           `json:"inQueue"`
	HealthReport       []HealthReport `json:"healthReport"`
	Builds             []BuildRef     `json:"builds"`
	LastBuild          *BuildRef      `json:"lastBuild"`
	LastCompletedBuild *BuildRef      `json:"lastCompletedBuild"`
	LastFailedBuild    *BuildRef      `json:"lastFailedBuild"`
}

func (c *Client) ListJobs(ctx context.Context, folder string) ([]JobSummary, error) {
	path := JobPath(folder) + "/api/json?tree=jobs[name,url,color]"
	if folder == "" {
		path = "/api/json?tree=jobs[name,url,color]"
	}
	var resp struct {
		Jobs []JobSummary `json:"jobs"`
	}
	if _, err := c.GET(ctx, path, &resp); err != nil {
		return nil, fmt.Errorf("list jobs: %w", err)
	}
	return resp.Jobs, nil
}

func (c *Client) GetJob(ctx context.Context, name string) (Job, error) {
	var j Job
	if _, err := c.GET(ctx, JobPath(name)+"/api/json", &j); err != nil {
		return Job{}, fmt.Errorf("get job %s: %w", name, err)
	}
	return j, nil
}

func (c *Client) GetJobConfig(ctx context.Context, name string) ([]byte, error) {
	body, err := c.GET(ctx, JobPath(name)+"/config.xml", nil)
	if err != nil {
		return nil, fmt.Errorf("get job config %s: %w", name, err)
	}
	return body, nil
}
```

- [ ] **Step 4: Run, PASS**

Run: `go test ./internal/jenkins -v`

- [ ] **Step 5: Write `cmd/ciq/jenkins_job.go`**

```go
package main

import (
	"fmt"
	"os"

	"github.com/kagent-dev/ciq/internal/output"
	"github.com/spf13/cobra"
)

func newJenkinsJobCmd() *cobra.Command {
	c := &cobra.Command{Use: "job", Short: "Inspect Jenkins jobs"}

	listFolder := ""
	list := &cobra.Command{
		Use:   "list",
		Short: "List jobs (optionally inside a folder)",
		RunE: func(cmd *cobra.Command, args []string) error {
			cl, ctx, err := newJenkinsClient(cmd)
			if err != nil {
				return err
			}
			jobs, err := cl.ListJobs(ctx, listFolder)
			if err != nil {
				return err
			}
			rows := make([]map[string]any, 0, len(jobs))
			for _, j := range jobs {
				rows = append(rows, map[string]any{"name": j.Name, "color": j.Color, "url": j.URL})
			}
			return output.Render(os.Stdout, output.Detect(stdoutIsTTY(), rf.format), rows)
		},
	}
	list.Flags().StringVar(&listFolder, "folder", "", "folder path (e.g. team/service)")

	get := &cobra.Command{
		Use:   "get <name>",
		Short: "Get a job's metadata",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl, ctx, err := newJenkinsClient(cmd)
			if err != nil {
				return err
			}
			j, err := cl.GetJob(ctx, args[0])
			if err != nil {
				return err
			}
			// For non-JSON formats flatten to a single-row table.
			f := output.Detect(stdoutIsTTY(), rf.format)
			if f == output.FormatJSON {
				return output.Render(os.Stdout, f, j)
			}
			row := map[string]any{"name": j.Name, "color": j.Color, "buildable": j.Buildable, "inQueue": j.InQueue}
			if j.LastBuild != nil {
				row["lastBuild"] = j.LastBuild.Number
			}
			return output.Render(os.Stdout, f, []map[string]any{row})
		},
	}

	cfg := &cobra.Command{
		Use:   "config <name>",
		Short: "Print a job's config.xml",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl, ctx, err := newJenkinsClient(cmd)
			if err != nil {
				return err
			}
			b, err := cl.GetJobConfig(ctx, args[0])
			if err != nil {
				return err
			}
			_, err = fmt.Fprintln(os.Stdout, string(b))
			return err
		},
	}

	c.AddCommand(list, get, cfg)
	return c
}
```

- [ ] **Step 6: Register in `cmd/ciq/jenkins.go`** — add `c.AddCommand(newJenkinsJobCmd())` next to whoami.

- [ ] **Step 7: Build, smoke-test help**

Run: `go build ./cmd/ciq && ./ciq jenkins job --help && ./ciq jenkins job list --help && ./ciq jenkins job get --help && ./ciq jenkins job config --help`

- [ ] **Step 8: Commit**

```bash
git add internal/jenkins/jobs.go internal/jenkins/jobs_test.go cmd/ciq/jenkins_job.go cmd/ciq/jenkins.go
git commit -s -m "feat(jenkins): job list, get, and config subcommands"
```

---

## Task 8: Builds Subcommands (`list`, `get`)

**Files:**
- Create: `E:\claudecode\mykagent\ciq\internal\jenkins\builds.go`
- Create: `E:\claudecode\mykagent\ciq\internal\jenkins\builds_test.go`
- Create: `E:\claudecode\mykagent\ciq\cmd\ciq\jenkins_build.go`
- Modify: `E:\claudecode\mykagent\ciq\cmd\ciq\jenkins.go` — register `build`.

**Interfaces:**
- Produces:
  ```go
  type Build struct {
      Number      int
      Result      string  // SUCCESS, FAILURE, UNSTABLE, ABORTED, null while running
      Building    bool
      Duration    int64   // ms
      Timestamp   int64   // unix ms
      URL         string
      DisplayName string
      Causes      []Cause `json:"-"` // populated via actions[*].causes flattening
  }
  type Cause struct { ShortDescription string; UserID, UserName string }

  func (c *Client) ListBuilds(ctx context.Context, jobName string, limit int) ([]Build, error)
  func (c *Client) GetBuild(ctx context.Context, jobName string, number int) (Build, error)
  ```

- [ ] **Step 1: Failing tests** — `internal/jenkins/builds_test.go`

```go
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
```

- [ ] **Step 2: Run, FAIL**

Run: `go test ./internal/jenkins -run TestListBuilds -run TestGetBuild -v`

- [ ] **Step 3: Implement** — `internal/jenkins/builds.go`

```go
package jenkins

import (
	"context"
	"fmt"
)

type Build struct {
	Number      int    `json:"number"`
	Result      string `json:"result"`
	Building    bool   `json:"building"`
	Duration    int64  `json:"duration"`
	Timestamp   int64  `json:"timestamp"`
	URL         string `json:"url"`
	DisplayName string `json:"displayName"`
}

func (c *Client) ListBuilds(ctx context.Context, jobName string, limit int) ([]Build, error) {
	if limit <= 0 {
		limit = 20
	}
	path := fmt.Sprintf("%s/api/json?tree=builds[number,result,building,duration,timestamp,url,displayName]{0,%d}",
		JobPath(jobName), limit)
	var resp struct {
		Builds []Build `json:"builds"`
	}
	if _, err := c.GET(ctx, path, &resp); err != nil {
		return nil, fmt.Errorf("list builds for %s: %w", jobName, err)
	}
	return resp.Builds, nil
}

func (c *Client) GetBuild(ctx context.Context, jobName string, number int) (Build, error) {
	path := fmt.Sprintf("%s/%d/api/json", JobPath(jobName), number)
	var b Build
	if _, err := c.GET(ctx, path, &b); err != nil {
		return Build{}, fmt.Errorf("get build %s #%d: %w", jobName, number, err)
	}
	return b, nil
}
```

- [ ] **Step 4: Run, PASS**

Run: `go test ./internal/jenkins -v`

- [ ] **Step 5: Write `cmd/ciq/jenkins_build.go`**

```go
package main

import (
	"os"
	"strconv"

	"github.com/kagent-dev/ciq/internal/output"
	"github.com/spf13/cobra"
)

func newJenkinsBuildCmd() *cobra.Command {
	c := &cobra.Command{Use: "build", Short: "Inspect Jenkins builds"}

	var limit int
	list := &cobra.Command{
		Use:   "list <job>",
		Short: "List recent builds for a job",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl, ctx, err := newJenkinsClient(cmd)
			if err != nil {
				return err
			}
			bs, err := cl.ListBuilds(ctx, args[0], limit)
			if err != nil {
				return err
			}
			rows := make([]map[string]any, 0, len(bs))
			for _, b := range bs {
				rows = append(rows, map[string]any{
					"number": b.Number, "result": b.Result, "building": b.Building,
					"duration_ms": b.Duration, "url": b.URL,
				})
			}
			return output.Render(os.Stdout, output.Detect(stdoutIsTTY(), rf.format), rows)
		},
	}
	list.Flags().IntVar(&limit, "limit", 20, "max number of builds to list")

	get := &cobra.Command{
		Use:   "get <job> <number>",
		Short: "Get build metadata",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			n, err := strconv.Atoi(args[1])
			if err != nil {
				return err
			}
			cl, ctx, err := newJenkinsClient(cmd)
			if err != nil {
				return err
			}
			b, err := cl.GetBuild(ctx, args[0], n)
			if err != nil {
				return err
			}
			return output.Render(os.Stdout, output.Detect(stdoutIsTTY(), rf.format), b)
		},
	}

	c.AddCommand(list, get)
	return c
}
```

- [ ] **Step 6: Register in `jenkins.go`** — add `c.AddCommand(newJenkinsBuildCmd())`.

- [ ] **Step 7: Commit**

```bash
git add internal/jenkins/builds.go internal/jenkins/builds_test.go cmd/ciq/jenkins_build.go cmd/ciq/jenkins.go
git commit -s -m "feat(jenkins): build list and get subcommands"
```

---

## Task 9: Console Subcommand (`console <job> <num>` with `--tail` / `--full`)

**Files:**
- Create: `E:\claudecode\mykagent\ciq\internal\jenkins\console.go`
- Create: `E:\claudecode\mykagent\ciq\internal\jenkins\console_test.go`
- Create: `E:\claudecode\mykagent\ciq\cmd\ciq\jenkins_console.go`
- Modify: `cmd/ciq/jenkins.go`.

**Interfaces:**
- Produces:
  ```go
  func (c *Client) GetConsole(ctx context.Context, jobName string, number int) ([]byte, error)
  // helpers used by CLI:
  func LastNLines(b []byte, n int) []byte
  ```

**Design note:** Jenkins exposes `/job/<name>/<num>/consoleText` (plain text) and `/job/<name>/<num>/logText/progressiveText?start=<offset>` (streaming). We use the simpler `consoleText` and do tail in-process; this matches the agent use-case (analyze failed build → grab last 200 lines).

- [ ] **Step 1: Failing test** — `internal/jenkins/console_test.go`

```go
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
```

- [ ] **Step 2: Run, FAIL**

Run: `go test ./internal/jenkins -run TestGetConsole -run TestLastNLines -v`

- [ ] **Step 3: Implement** — `internal/jenkins/console.go`

```go
package jenkins

import (
	"bytes"
	"context"
	"fmt"
)

func (c *Client) GetConsole(ctx context.Context, jobName string, number int) ([]byte, error) {
	path := fmt.Sprintf("%s/%d/consoleText", JobPath(jobName), number)
	b, err := c.GET(ctx, path, nil)
	if err != nil {
		return nil, fmt.Errorf("get console %s #%d: %w", jobName, number, err)
	}
	return b, nil
}

// LastNLines returns the last n newline-terminated (or trailing) lines of b.
func LastNLines(b []byte, n int) []byte {
	if n <= 0 {
		return nil
	}
	count := 0
	for i := len(b) - 1; i >= 0; i-- {
		if b[i] == '\n' && i != len(b)-1 {
			count++
			if count == n {
				return b[i+1:]
			}
		}
	}
	// fewer than n lines exist
	return bytes.Clone(b)
}
```

- [ ] **Step 4: Run, PASS**

Run: `go test ./internal/jenkins -v`

- [ ] **Step 5: Write `cmd/ciq/jenkins_console.go`**

```go
package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/kagent-dev/ciq/internal/jenkins"
	"github.com/spf13/cobra"
)

func newJenkinsConsoleCmd() *cobra.Command {
	var tail int
	var full bool
	c := &cobra.Command{
		Use:   "console <job> <number>",
		Short: "Print build console log (default: last 200 lines)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			n, err := strconv.Atoi(args[1])
			if err != nil {
				return err
			}
			cl, ctx, err := newJenkinsClient(cmd)
			if err != nil {
				return err
			}
			body, err := cl.GetConsole(ctx, args[0], n)
			if err != nil {
				return err
			}
			out := body
			if !full {
				out = jenkins.LastNLines(body, tail)
				if len(out) < len(body) {
					fmt.Fprintf(os.Stderr, "(showing last %d lines of %d bytes; use --full for everything)\n", tail, len(body))
				}
			}
			_, err = os.Stdout.Write(out)
			return err
		},
	}
	c.Flags().IntVar(&tail, "tail", 200, "show last N lines")
	c.Flags().BoolVar(&full, "full", false, "print the entire console log")
	return c
}
```

- [ ] **Step 6: Register in `jenkins.go`** — add `c.AddCommand(newJenkinsConsoleCmd())`.

- [ ] **Step 7: Commit**

```bash
git add internal/jenkins/console.go internal/jenkins/console_test.go cmd/ciq/jenkins_console.go cmd/ciq/jenkins.go
git commit -s -m "feat(jenkins): console log subcommand with tail/full"
```

---

## Task 10: Clouds Subcommand (`cloud list`, `cloud get`)

**Files:**
- Create: `E:\claudecode\mykagent\ciq\internal\jenkins\clouds.go`
- Create: `E:\claudecode\mykagent\ciq\internal\jenkins\clouds_test.go`
- Create: `E:\claudecode\mykagent\ciq\cmd\ciq\jenkins_cloud.go`
- Modify: `cmd/ciq/jenkins.go`.

**Design note:** Jenkins kubernetes-plugin exposes cloud config under `/manage/cloud/<name>/config.xml`. There is no clean list-all JSON endpoint, so we list by parsing the global `config.xml` for `<clouds>` block. Tests assert against a mock returning a known config.xml.

**Interfaces:**
- Produces:
  ```go
  type Cloud struct { Name, Kind string } // Kind e.g. "kubernetes"
  func (c *Client) ListClouds(ctx context.Context) ([]Cloud, error)
  func (c *Client) GetCloudConfig(ctx context.Context, name string) ([]byte, error)
  ```

- [ ] **Step 1: Failing tests** — `internal/jenkins/clouds_test.go`

```go
package jenkins

import (
	"context"
	"net/http"
	"strings"
	"testing"
)

const sampleRootConfig = `<?xml version='1.1' encoding='UTF-8'?>
<hudson>
  <clouds>
    <org.csanchez.jenkins.plugins.kubernetes.KubernetesCloud plugin="kubernetes@4258.v1234">
      <name>kubernetes</name>
    </org.csanchez.jenkins.plugins.kubernetes.KubernetesCloud>
    <io.jenkins.plugins.ec2.EC2Cloud>
      <name>ec2-builders</name>
    </io.jenkins.plugins.ec2.EC2Cloud>
  </clouds>
</hudson>`

func TestListClouds(t *testing.T) {
	c, stop := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/config.xml" {
			t.Errorf("path=%s", r.URL.Path)
		}
		_, _ = w.Write([]byte(sampleRootConfig))
	})
	defer stop()
	clouds, err := c.ListClouds(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(clouds) != 2 {
		t.Fatalf("got %d clouds: %+v", len(clouds), clouds)
	}
	if clouds[0].Name != "kubernetes" || !strings.Contains(clouds[0].Kind, "Kubernetes") {
		t.Errorf("clouds[0]=%+v", clouds[0])
	}
	if clouds[1].Name != "ec2-builders" {
		t.Errorf("clouds[1]=%+v", clouds[1])
	}
}

func TestGetCloudConfig(t *testing.T) {
	c, stop := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/manage/cloud/kubernetes/config.xml" {
			t.Errorf("path=%s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`<KubernetesCloud><name>kubernetes</name></KubernetesCloud>`))
	})
	defer stop()
	b, err := c.GetCloudConfig(context.Background(), "kubernetes")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "<KubernetesCloud>") {
		t.Errorf("got %s", string(b))
	}
}
```

- [ ] **Step 2: Run, FAIL**

Run: `go test ./internal/jenkins -run TestListClouds -run TestGetCloudConfig -v`

- [ ] **Step 3: Implement** — `internal/jenkins/clouds.go`

```go
package jenkins

import (
	"context"
	"encoding/xml"
	"fmt"
	"strings"
)

type Cloud struct {
	Name string `json:"name"`
	Kind string `json:"kind"` // simplified class name, e.g. "KubernetesCloud"
}

func (c *Client) ListClouds(ctx context.Context) ([]Cloud, error) {
	body, err := c.GET(ctx, "/config.xml", nil)
	if err != nil {
		return nil, fmt.Errorf("list clouds (root config): %w", err)
	}
	return parseCloudList(body)
}

func (c *Client) GetCloudConfig(ctx context.Context, name string) ([]byte, error) {
	body, err := c.GET(ctx, "/manage/cloud/"+name+"/config.xml", nil)
	if err != nil {
		return nil, fmt.Errorf("get cloud %s config: %w", name, err)
	}
	return body, nil
}

// parseCloudList walks <clouds> children, extracting <name>...</name> from each.
func parseCloudList(body []byte) ([]Cloud, error) {
	dec := xml.NewDecoder(strings.NewReader(string(body)))
	var clouds []Cloud
	inClouds := false
	var current *Cloud
	for {
		tok, err := dec.Token()
		if err != nil {
			break
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if t.Name.Local == "clouds" {
				inClouds = true
				continue
			}
			if inClouds && current == nil {
				// First child element inside <clouds> = a cloud entry
				current = &Cloud{Kind: simpleKind(t.Name.Local)}
				continue
			}
			if current != nil && t.Name.Local == "name" {
				var name string
				if err := dec.DecodeElement(&name, &t); err == nil {
					current.Name = name
				}
			}
		case xml.EndElement:
			if t.Name.Local == "clouds" {
				inClouds = false
			}
			if current != nil && t.Name.Local != "name" && simpleKind(t.Name.Local) == current.Kind {
				clouds = append(clouds, *current)
				current = nil
			}
		}
	}
	return clouds, nil
}

// simpleKind extracts the trailing class name (after the last dot).
func simpleKind(full string) string {
	if i := strings.LastIndex(full, "."); i >= 0 {
		return full[i+1:]
	}
	return full
}
```

- [ ] **Step 4: Run, PASS**

Run: `go test ./internal/jenkins -v`

- [ ] **Step 5: Write `cmd/ciq/jenkins_cloud.go`**

```go
package main

import (
	"fmt"
	"os"

	"github.com/kagent-dev/ciq/internal/output"
	"github.com/spf13/cobra"
)

func newJenkinsCloudCmd() *cobra.Command {
	c := &cobra.Command{Use: "cloud", Short: "Inspect Jenkins clouds (K8s/EC2/etc.)"}

	list := &cobra.Command{
		Use: "list", Short: "List configured clouds",
		RunE: func(cmd *cobra.Command, args []string) error {
			cl, ctx, err := newJenkinsClient(cmd)
			if err != nil {
				return err
			}
			cs, err := cl.ListClouds(ctx)
			if err != nil {
				return err
			}
			rows := make([]map[string]any, 0, len(cs))
			for _, x := range cs {
				rows = append(rows, map[string]any{"name": x.Name, "kind": x.Kind})
			}
			return output.Render(os.Stdout, output.Detect(stdoutIsTTY(), rf.format), rows)
		},
	}

	get := &cobra.Command{
		Use: "get <name>", Short: "Print a cloud's config.xml",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl, ctx, err := newJenkinsClient(cmd)
			if err != nil {
				return err
			}
			b, err := cl.GetCloudConfig(ctx, args[0])
			if err != nil {
				return err
			}
			_, err = fmt.Fprintln(os.Stdout, string(b))
			return err
		},
	}

	c.AddCommand(list, get)
	return c
}
```

- [ ] **Step 6: Register in `jenkins.go`** — add `c.AddCommand(newJenkinsCloudCmd())`.

- [ ] **Step 7: Commit**

```bash
git add internal/jenkins/clouds.go internal/jenkins/clouds_test.go cmd/ciq/jenkins_cloud.go cmd/ciq/jenkins.go
git commit -s -m "feat(jenkins): cloud list and get subcommands"
```

---

## Task 11: Queue Subcommand (`queue list`)

**Files:**
- Create: `E:\claudecode\mykagent\ciq\internal\jenkins\queue.go`
- Create: `E:\claudecode\mykagent\ciq\internal\jenkins\queue_test.go`
- Create: `E:\claudecode\mykagent\ciq\cmd\ciq\jenkins_queue.go`
- Modify: `cmd/ciq/jenkins.go`.

**Interfaces:**
- Produces:
  ```go
  type QueueItem struct {
      ID    int64
      Why   string
      Stuck bool
      Task  struct { Name, URL string }
      InQueueSince int64
  }
  func (c *Client) ListQueue(ctx context.Context) ([]QueueItem, error)
  ```

- [ ] **Step 1: Failing test**

```go
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
```

- [ ] **Step 2: Run, FAIL**

- [ ] **Step 3: Implement** — `internal/jenkins/queue.go`

```go
package jenkins

import (
	"context"
	"fmt"
)

type QueueItem struct {
	ID    int64  `json:"id"`
	Why   string `json:"why"`
	Stuck bool   `json:"stuck"`
	Task  struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"task"`
	InQueueSince int64 `json:"inQueueSince"`
}

func (c *Client) ListQueue(ctx context.Context) ([]QueueItem, error) {
	var resp struct {
		Items []QueueItem `json:"items"`
	}
	if _, err := c.GET(ctx, "/queue/api/json", &resp); err != nil {
		return nil, fmt.Errorf("list queue: %w", err)
	}
	return resp.Items, nil
}
```

- [ ] **Step 4: Run, PASS**

- [ ] **Step 5: Write `cmd/ciq/jenkins_queue.go`**

```go
package main

import (
	"os"

	"github.com/kagent-dev/ciq/internal/output"
	"github.com/spf13/cobra"
)

func newJenkinsQueueCmd() *cobra.Command {
	c := &cobra.Command{Use: "queue", Short: "Inspect Jenkins build queue"}
	list := &cobra.Command{
		Use: "list", Short: "List queued items",
		RunE: func(cmd *cobra.Command, args []string) error {
			cl, ctx, err := newJenkinsClient(cmd)
			if err != nil {
				return err
			}
			items, err := cl.ListQueue(ctx)
			if err != nil {
				return err
			}
			rows := make([]map[string]any, 0, len(items))
			for _, q := range items {
				rows = append(rows, map[string]any{
					"id": q.ID, "task": q.Task.Name, "why": q.Why, "stuck": q.Stuck,
				})
			}
			return output.Render(os.Stdout, output.Detect(stdoutIsTTY(), rf.format), rows)
		},
	}
	c.AddCommand(list)
	return c
}
```

- [ ] **Step 6: Register in `jenkins.go`** — add `c.AddCommand(newJenkinsQueueCmd())`.

- [ ] **Step 7: Commit**

```bash
git add internal/jenkins/queue.go internal/jenkins/queue_test.go cmd/ciq/jenkins_queue.go cmd/ciq/jenkins.go
git commit -s -m "feat(jenkins): queue list subcommand"
```

---

## Task 12: Skill Markdown (`skills/jenkins.md`)

**Files:**
- Create: `E:\claudecode\mykagent\ciq\skills\jenkins.md`

This is the artifact any agent loads. Self-contained, no kagent-specific references in body.

- [ ] **Step 1: Write `skills/jenkins.md`**

````markdown
---
name: jenkins
description: Query Jenkins read-only (jobs, builds, console, clouds, queue) via the ciq CLI. Safe to use during incident triage; never mutates state.
---

# Jenkins (read-only) via `ciq`

You have access to a CLI named `ciq` that wraps the Jenkins REST API. **All operations are GET-only — this binary cannot trigger, abort, or modify any Jenkins object.** Treat this as a safe inspection tool.

## When to use

- Diagnose a failed build (read console log, inspect build metadata, fetch job config).
- Inventory jobs across folders.
- Inspect Kubernetes / EC2 cloud configurations (e.g., investigate why pod-templated agents are not scheduling).
- Inspect the build queue.

## Authentication

Already configured. The user has provided credentials in `~/.config/ciq/credentials.yaml`. If multiple environments are available, switch with `--context <name>` (e.g. `--context staging`). Never ask the user for credentials.

## Output

Add `--format json` for structured output you can parse. Without that flag the output is a human table; if stdout is not a terminal, `ciq` defaults to JSON automatically.

## Command catalog

| What you need | Command |
|---|---|
| Confirm auth works | `ciq jenkins whoami` |
| List jobs (top-level) | `ciq jenkins job list` |
| List jobs in a folder | `ciq jenkins job list --folder team/service` |
| Job metadata + last build | `ciq jenkins job get <name>` |
| Full job config.xml (pipeline def, params, triggers) | `ciq jenkins job config <name>` |
| List recent builds | `ciq jenkins build list <name> --limit 10` |
| One build's metadata (result, duration, cause) | `ciq jenkins build get <name> <number>` |
| Last 200 lines of console log | `ciq jenkins console <name> <number>` |
| Full console log | `ciq jenkins console <name> <number> --full` |
| Last 50 lines (smaller context) | `ciq jenkins console <name> <number> --tail 50` |
| List configured clouds | `ciq jenkins cloud list` |
| One cloud's config.xml (e.g. k8s pod templates) | `ciq jenkins cloud get <name>` |
| What's queued right now | `ciq jenkins queue list` |

## Job names with folders

Jenkins folders nest jobs. Pass the full path with slashes:

```
ciq jenkins job get org/team/service-main
ciq jenkins console org/team/service-main 42
```

## Triage playbook (build failed)

1. `ciq jenkins build get <job> <num> --format json` — get `result`, `duration`, `url`.
2. `ciq jenkins console <job> <num> --tail 300` — usually enough to spot the failure line.
3. If the failure mentions a missing agent / pod, run `ciq jenkins cloud list` and `ciq jenkins cloud get <cloud>` to inspect Kubernetes pod templates and limits.
4. If the failure is config-shaped (wrong branch, missing parameter), run `ciq jenkins job config <job>` to view the pipeline definition.
5. If many builds are queued, `ciq jenkins queue list` and look for `stuck:true` items.

## Hard rules

- **Never** suggest a `ciq` command that would mutate Jenkins. The CLI does not support it. If the user explicitly asks to retry/abort/replay, tell them to use the Jenkins UI or their `gh`/`glab`/`kubectl` equivalents — do not invoke `curl` to Jenkins yourself.
- **Never** print or echo the credentials file. The user can manage it.
- Console logs can be huge — default to `--tail 200` and only escalate to `--full` if you've located the failure region.
````

- [ ] **Step 2: Commit**

```bash
git add skills/jenkins.md
git commit -s -m "docs(skill): jenkins read-only skill for ai agents"
```

---

## Task 13: README and `docs/architecture.md` Finalization

**Files:**
- Modify: `E:\claudecode\mykagent\ciq\README.md` — replace placeholder body.
- Create: `E:\claudecode\mykagent\ciq\docs\architecture.md`

- [ ] **Step 1: Write final `README.md`**

```markdown
# ciq — AI-agent-friendly CI inspection CLI

`ciq` is a small Go CLI that lets AI agents (and humans) query CI/CD systems safely.
Jenkins is the first backend; GitLab CI and GitHub Actions are planned.

## Why

Agents that operate against Jenkins via raw `curl` are fragile (CSRF crumbs, folder paths, XML)
and unsafe (they can construct POSTs to trigger or abort jobs). Wrapping the API in a
purpose-built CLI gives us three things:

1. **Compile-time read-only guarantee.** This build's HTTP client has no `POST` / `PUT` / `DELETE`
   helper. No subcommand can mutate Jenkins.
2. **LLM-friendly output.** Every command supports `--format json|table|md`; output is
   structured and stable.
3. **Centralized auth.** The token lives in `~/.config/ciq/credentials.yaml` and is never
   echoed into command output, error messages, or LLM context.

## Install

```
go install github.com/kagent-dev/ciq/cmd/ciq@latest
```

Pre-built binaries: see Releases.

## Configure

Copy `examples/credentials.yaml.example` to `~/.config/ciq/credentials.yaml` and fill in
a Jenkins API token (generate it under **People → Your User → Configure → API Token**).

## Use

```
ciq jenkins whoami
ciq jenkins job list
ciq jenkins job get team/service/main
ciq jenkins build get team/service/main 42
ciq jenkins console team/service/main 42 --tail 200
ciq jenkins cloud list
ciq jenkins cloud get kubernetes
ciq jenkins queue list
```

Multi-environment? `--context staging`. Pipe to `jq`? Add `--format json`.

## Use with AI agents

Ship `skills/jenkins.md` alongside the binary. Any agent framework that loads skill
markdown — kagent (`Agent.spec.skills.gitRefs`), Claude Code (`~/.claude/skills/`),
Cursor — can drop it in and start triaging Jenkins failures.

## Roadmap

- GitLab CI backend (`ciq gitlab ...`)
- GitHub Actions backend (`ciq github ...`)
- Argo Workflows, Tekton
- Optional mutating build behind `--enable-mutations` Go build tag (separate binary)

## License

Apache 2.0.
```

- [ ] **Step 2: Write `docs/architecture.md`**

```markdown
# Architecture

```
┌──────────────────┐
│  cmd/ciq         │  cobra subcommands; no HTTP knowledge
│  (one file per   │
│   resource)      │
└────────┬─────────┘
         │ calls
         ▼
┌──────────────────┐
│ internal/jenkins │  one file per resource (jobs, builds, console, clouds, queue)
│                  │  every method has an httptest-backed table-driven test
└────────┬─────────┘
         │ uses
         ▼
┌──────────────────┐
│ internal/jenkins │  client.go: HTTP client struct, GET-only, basic auth,
│  client.go       │            token scrubbing on errors
└────────┬─────────┘
         │ uses
         ▼
┌──────────────────┐
│ internal/config  │  YAML-based credentials loader, context switching
└──────────────────┘

┌──────────────────┐
│ internal/output  │  json/table/md formatters, tty detection
└──────────────────┘
```

## Read-only is structural, not a flag

`internal/jenkins/client.go` exposes exactly one method: `GET(ctx, path, into)`. There is no
`POST` helper to misuse. Adding mutation in the future requires editing this file, which is
where the read-only review boundary lives.

## Adding a new backend

1. Add a new package under `internal/<name>/` mirroring `internal/jenkins/`.
2. Add a `cmd/ciq/<name>.go` parent command and one subcommand file per resource.
3. Each new resource type needs an `httptest`-backed test before any CLI surface.
4. Add a `skills/<name>.md` describing the agent-facing surface.
```

- [ ] **Step 3: Commit**

```bash
git add README.md docs/architecture.md
git commit -s -m "docs: finalize readme and architecture"
```

---

## Task 14: kagent PR Materials — `helm/agents/jenkins-triage/`

This task produces the directory we will copy into a kagent fork to open a PR. The kagent
target path is `helm/agents/jenkins-triage/`. We stage it locally under
`E:\claudecode\mykagent\ciq\kagent-pr\helm-agents-jenkins-triage\`.

**Files:**
- Create: `E:\claudecode\mykagent\ciq\kagent-pr\helm-agents-jenkins-triage\Chart-template.yaml`
- Create: `E:\claudecode\mykagent\ciq\kagent-pr\helm-agents-jenkins-triage\values.yaml`
- Create: `E:\claudecode\mykagent\ciq\kagent-pr\helm-agents-jenkins-triage\templates\_helpers.tpl`
- Create: `E:\claudecode\mykagent\ciq\kagent-pr\helm-agents-jenkins-triage\templates\agent.yaml`
- Create: `E:\claudecode\mykagent\ciq\examples\jenkins-triage-agent.yaml`
- Create: `E:\claudecode\mykagent\ciq\docs\kagent-pr-checklist.md`

**Reference shape:** Patterned after `kagent/helm/agents/observability/`, which I inspected
during planning: `Chart-template.yaml` + `values.yaml` + `templates/{_helpers.tpl, agent.yaml}`.

- [ ] **Step 1: `Chart-template.yaml`**

```yaml
apiVersion: v2
name: jenkins-triage
description: Jenkins read-only triage agent (uses the ciq CLI + jenkins skill).
type: application
# Filled in by kagent's release tooling.
version: 0.0.0
appVersion: "0.1.0"
```

- [ ] **Step 2: `values.yaml`**

```yaml
namespace: kagent

agent:
  name: jenkins-triage-agent
  modelConfigRef:
    name: default-model-config
  description: |
    Diagnose Jenkins build failures read-only. Inspects job config, build metadata,
    console logs, cloud (Kubernetes) configuration, and the build queue.
  systemPrompt: |
    You are a Jenkins triage agent. You can only QUERY Jenkins — you cannot trigger,
    retry, or abort builds. When the user reports a failed build, fetch the console log
    (default tail 200), identify the failing step, and explain what likely went wrong.
    For environment issues, inspect the cloud config. If a fix requires a UI action,
    state that clearly; do not attempt mutations.

# The ciq skill is pulled from the public OSS repo by Git ref.
skills:
  gitRefs:
    - url: https://github.com/kagent-dev/ciq
      ref: v0.1.0
      path: skills/jenkins.md

# The ciq binary must be present in the agent runtime container. Two options:
#   1) Use a custom runtime image that has `ciq` baked in (recommended for production).
#   2) Mount the binary from a hostPath / initContainer (lab only).
# This chart assumes (1). Override `runtimeImage` to point at your image.
runtimeImage: ghcr.io/your-org/kagent-runtime-with-ciq:latest

# Mount a credentials secret containing ~/.config/ciq/credentials.yaml.
credentialsSecret:
  name: jenkins-ciq-credentials
  key: credentials.yaml
```

- [ ] **Step 3: `templates/_helpers.tpl`**

```yaml
{{- define "jenkins-triage.fullname" -}}
{{ .Values.agent.name }}
{{- end -}}
```

- [ ] **Step 4: `templates/agent.yaml`**

```yaml
apiVersion: kagent.dev/v1alpha2
kind: Agent
metadata:
  name: {{ include "jenkins-triage.fullname" . }}
  namespace: {{ .Values.namespace }}
spec:
  description: {{ .Values.agent.description | quote }}
  systemMessage: |
    {{- .Values.agent.systemPrompt | nindent 4 }}
  modelConfig:
    name: {{ .Values.agent.modelConfigRef.name }}
  skills:
    gitRefs:
      {{- range .Values.skills.gitRefs }}
      - url: {{ .url | quote }}
        ref: {{ .ref | quote }}
        path: {{ .path | quote }}
      {{- end }}
  runtime:
    image: {{ .Values.runtimeImage | quote }}
    env:
      - name: HOME
        value: /home/agent
    volumeMounts:
      - name: ciq-credentials
        mountPath: /home/agent/.config/ciq
        readOnly: true
  volumes:
    - name: ciq-credentials
      secret:
        secretName: {{ .Values.credentialsSecret.name }}
        items:
          - key: {{ .Values.credentialsSecret.key }}
            path: credentials.yaml
```

> **NOTE for the reviewer:** The exact CRD fields (`runtime`, `volumes`) need to be reconciled
> against the current kagent `v1alpha2` Agent CRD before submission. The vendoring
> checklist (next file) covers this.

- [ ] **Step 5: `examples/jenkins-triage-agent.yaml`** (standalone CRD, no Helm)

```yaml
apiVersion: kagent.dev/v1alpha2
kind: Agent
metadata:
  name: jenkins-triage-agent
  namespace: kagent
spec:
  description: Read-only Jenkins triage.
  systemMessage: |
    You are a Jenkins triage agent. Use ciq jenkins commands. Read-only only.
  modelConfig:
    name: default-model-config
  skills:
    gitRefs:
      - url: https://github.com/kagent-dev/ciq
        ref: v0.1.0
        path: skills/jenkins.md
```

- [ ] **Step 6: `docs/kagent-pr-checklist.md`** — the operational guide to actually open the PR.

```markdown
# kagent PR Checklist

Goal: land `helm/agents/jenkins-triage/` plus an example Agent YAML in the kagent
repository as a PR.

## Before submitting

1. **Verify the CRD shape.** Open `go/api/v1alpha2/agent_types.go` in the kagent repo and
   confirm:
   - `Agent.spec.skills.gitRefs[*]` exists and accepts `url`, `ref`, `path` fields.
   - `Agent.spec.runtime` exposes `image`, `env`, `volumeMounts`.
   - `Agent.spec.volumes` accepts a `[]corev1.Volume`-shaped slice.
   If any field is named differently, adjust `templates/agent.yaml` accordingly.
2. **Match the existing chart pattern.** Compare against `helm/agents/observability/` for:
   - `Chart-template.yaml` keys.
   - `_helpers.tpl` naming convention.
   - `values.yaml` key names and indentation.
3. **Lint locally.** Run `helm lint helm/agents/jenkins-triage` from the kagent root.

## Vendoring

From the `ciq` repo root, with `KAGENT_DIR` pointing at your kagent fork:

```
make vendor-to-kagent KAGENT_DIR=$HOME/code/kagent
```

This copies `kagent-pr/helm-agents-jenkins-triage/` into
`$KAGENT_DIR/helm/agents/jenkins-triage/`.

## Commit + PR

Inside the kagent fork:

```
git checkout -b feat/jenkins-triage-agent
git add helm/agents/jenkins-triage
git commit -s -m "feat: add jenkins-triage agent (read-only via ciq)"
git push origin feat/jenkins-triage-agent
```

PR description template:

```
## What
Adds a new built-in agent `jenkins-triage` for diagnosing Jenkins build failures
read-only.

## How
Uses the open-source `ciq` CLI (https://github.com/kagent-dev/ciq) for all Jenkins
queries. Loads the `jenkins` skill via `Agent.spec.skills.gitRefs`. Read-only by
construction (the `ciq` build has no mutating subcommands).

## Test plan
- [ ] `helm lint helm/agents/jenkins-triage`
- [ ] Deploy on a Kind cluster with a Jenkins instance reachable; create
      `jenkins-ciq-credentials` secret; verify the agent answers
      "list jobs in folder team/service" correctly.
- [ ] Verify no POST/PUT/DELETE requests appear in Jenkins access logs during a triage
      conversation.
```
```

- [ ] **Step 7: Commit**

```bash
git add kagent-pr examples/jenkins-triage-agent.yaml docs/kagent-pr-checklist.md
git commit -s -m "feat(kagent-pr): helm chart + checklist for jenkins-triage agent"
```

---

## Task 15: Jenkins Shared Library Guide (`kagentAnalyze()`)

**Files:**
- Create: `E:\claudecode\mykagent\jenkins-shared-library-guide.md`

This document is **outside** the ciq project — it lives at the workspace root and is
**not** committed to either ciq or kagent. It is a standalone how-to for plugging
Jenkins post-failure into the kagent jenkins-triage agent.

- [ ] **Step 1: Write the guide**

````markdown
# Jenkins Shared Library: `kagentAnalyze()`

A Jenkins **Shared Library** (not a plugin) that exposes a single global function
`kagentAnalyze()` so every team's Jenkinsfile can opt into automatic AI triage on
failure:

```groovy
pipeline {
  agent any
  stages { /* ... */ }
  post {
    failure { kagentAnalyze() }
  }
}
```

The library has zero compile step and no Update-Center release process. You version-control
it in a Git repo and Jenkins loads it dynamically.

## Architecture

```
┌─────────────┐  failure   ┌────────────────┐  HTTP POST   ┌────────────────────┐
│  Jenkinsfile│ ─────────▶ │ kagentAnalyze()│ ──────────▶  │ jenkins-triage     │
│  post block │            │ (shared lib)   │              │ agent (A2A endpt)  │
└─────────────┘            └────────────────┘              └────────┬───────────┘
                                                                   │
                                              ┌────────────────────┴─────────────────┐
                                              │ Agent uses `ciq jenkins …` to fetch  │
                                              │ console, build meta, cloud config,   │
                                              │ then writes summary back to:         │
                                              │   - build description, and/or         │
                                              │   - Slack via kagent-slack-mcp        │
                                              └──────────────────────────────────────┘
```

## Repo layout

Create a new Git repo (e.g. `jenkins-shared-libs`) with:

```
jenkins-shared-libs/
├── vars/
│   └── kagentAnalyze.groovy     # global step
├── src/
│   └── io/kagent/jenkins/
│       └── KagentClient.groovy  # HTTP client class
└── README.md
```

### `vars/kagentAnalyze.groovy`

```groovy
// vars/kagentAnalyze.groovy
import io.kagent.jenkins.KagentClient

/**
 * Triggers a kagent jenkins-triage analysis for the current build.
 *
 * Usage in Jenkinsfile:
 *   post { failure { kagentAnalyze() } }
 *
 * Or with overrides:
 *   post { failure { kagentAnalyze(endpoint: 'https://kagent.internal/a2a/jenkins-triage', tail: 500) } }
 */
def call(Map args = [:]) {
    String endpoint = args.endpoint ?: env.KAGENT_TRIAGE_ENDPOINT
    Integer tail    = (args.tail ?: 200) as Integer
    String  context = args.context ?: env.KAGENT_TRIAGE_CONTEXT ?: 'prod'

    if (!endpoint) {
        echo 'kagentAnalyze: KAGENT_TRIAGE_ENDPOINT not configured, skipping'
        return
    }

    def payload = [
        job        : env.JOB_NAME,
        build      : env.BUILD_NUMBER as Integer,
        buildUrl   : env.BUILD_URL,
        result     : currentBuild.currentResult,
        durationMs : currentBuild.duration,
        gitCommit  : env.GIT_COMMIT,
        gitBranch  : env.GIT_BRANCH,
        cause      : currentBuild.getBuildCauses().collect { it.shortDescription }.join('; '),
        consoleTail: tail,
        ciqContext : context,
    ]

    def client = new KagentClient(this, endpoint)
    def summary = client.analyze(payload)

    if (summary) {
        // Write the agent's verdict onto the build page.
        currentBuild.description = (currentBuild.description ?: '') +
            "\n--- kagent triage ---\n" + summary
    }
}
```

### `src/io/kagent/jenkins/KagentClient.groovy`

```groovy
package io.kagent.jenkins

class KagentClient implements Serializable {
    private final Object steps
    private final String endpoint

    KagentClient(Object steps, String endpoint) {
        this.steps = steps
        this.endpoint = endpoint
    }

    String analyze(Map payload) {
        // Token managed via Jenkins credential `kagent-triage-token` (string credential).
        return steps.withCredentials([
            steps.string(credentialsId: 'kagent-triage-token', variable: 'KAGENT_TOKEN')
        ]) {
            def json = groovy.json.JsonOutput.toJson(payload)
            def resp = steps.sh(
                returnStdout: true,
                script: """
                    curl -sS -X POST '${endpoint}' \\
                      -H 'Authorization: Bearer ${steps.env.KAGENT_TOKEN}' \\
                      -H 'Content-Type: application/json' \\
                      --max-time 120 \\
                      --data '${json.replace("'", "'\\\\''")}'
                """
            ).trim()
            try {
                def parsed = new groovy.json.JsonSlurper().parseText(resp)
                return parsed.summary ?: parsed.result ?: resp
            } catch (Exception ignored) {
                return resp
            }
        }
    }
}
```

## Configure Jenkins to load the library

**Manage Jenkins → System → Global Pipeline Libraries**:

| Field | Value |
|---|---|
| Name | `kagent` |
| Default version | `main` (or a tag like `v1.0.0`) |
| Load implicitly | unchecked (require explicit `@Library`) |
| Allow default version override | checked |
| Retrieval method | Modern SCM → Git → repo URL of `jenkins-shared-libs` |
| Credentials | a read-only SSH key or PAT |

## Configure the kagent triage endpoint

Two **Manage Jenkins → System → Global properties → Environment variables**:

| Name | Value |
|---|---|
| `KAGENT_TRIAGE_ENDPOINT` | `https://kagent.internal/a2a/jenkins-triage` |
| `KAGENT_TRIAGE_CONTEXT` | `prod` |

And a **Manage Credentials → Global** secret:

| Kind | ID | Value |
|---|---|---|
| Secret text | `kagent-triage-token` | Bearer token your kagent A2A endpoint accepts |

## Use it in any Jenkinsfile

```groovy
@Library('kagent') _

pipeline {
  agent any
  stages {
    stage('Build') { steps { sh 'make' } }
    stage('Test')  { steps { sh 'make test' } }
  }
  post {
    failure {
      kagentAnalyze()             // defaults: tail=200, context=prod
    }
    unstable {
      kagentAnalyze(tail: 500)
    }
  }
}
```

## Manual trigger (no failure)

Add a parameterized build:

```groovy
parameters {
  booleanParam(name: 'KAGENT_ANALYZE', defaultValue: false, description: 'Run kagent triage at end')
}
post {
  always {
    script { if (params.KAGENT_ANALYZE) { kagentAnalyze(tail: 1000) } }
  }
}
```

Users get a "Build with parameters" button to opt-in mid-build.

## What kagent sees on the other end

The `jenkins-triage-agent` (from the kagent PR in `helm/agents/jenkins-triage/`) receives
the JSON payload as the initial user message, then autonomously runs:

```
ciq jenkins build get <job> <num>
ciq jenkins console <job> <num> --tail <consoleTail>
ciq jenkins job config <job>      # only if pipeline-level failure suspected
ciq jenkins cloud list            # only if agent/executor failures suspected
```

…and returns a `summary` field with the diagnosis, which `kagentAnalyze()` pastes into the
build description.

## Operational notes

- The shared library **shells out to `curl`**, which means the controller (or wherever
  `post` runs) needs network reach to your kagent service. In Jenkins-on-K8s setups, expose
  kagent via a ClusterIP and authorize the controller's ServiceAccount.
- **120-second timeout** is conservative; bump for very large logs but keep it bounded
  so a stuck agent does not wedge the pipeline.
- The library does **not** require Pipeline approval beyond the `withCredentials` and
  `sh` steps (already standard).
- Want richer payload? Add `manifestEntries`, `changeSet`, `previousBuildResult` — all
  available on `currentBuild`.

## Why a shared library and not a plugin

| | Shared Library | Plugin |
|---|---|---|
| Distribution | git push | Update Center + LTS compat matrix |
| UI changes | None (build description only) | Possible |
| Maintenance | low | high |
| Per-team override | Just edit the Jenkinsfile | Hard |

If you later need a "Replay with kagent" button on the build page, that's when a real
plugin starts to pay off.
````

- [ ] **Step 2: Done** (this file lives outside any repo, no commit needed).

---

## Self-Review

**Spec coverage:**
- ✅ "新建目录专门存放我们的这个cli+skill的项目，写README.md" → Tasks 1, 13.
- ✅ "做完了可以把skill+cli部分放到kagent仓库里然后直接给kagent的仓库推pr的程度" →
  Task 14 stages a Helm chart matching `helm/agents/observability/`'s structure, plus a
  `docs/kagent-pr-checklist.md` and `make vendor-to-kagent` target.
- ✅ "Jenkins Shared Library (`kagentAnalyze()`) 指导文档, 不需要推仓库" → Task 15 writes
  it to `E:\claudecode\mykagent\jenkins-shared-library-guide.md`, outside any repo.

**Placeholder scan:** All steps include exact code, exact commands, exact expected output.
The one explicit `NOTE for the reviewer` block in Task 14's agent.yaml flags a CRD-field
reconciliation step that is then formalized into a concrete checklist item in
`docs/kagent-pr-checklist.md` Step 6 ("Verify the CRD shape"). That is a real verification
task with concrete files to open, not a hidden TODO.

**Type consistency:**
- `config.Credentials` defined in Task 2 is consumed unchanged by `jenkins.New` in Task 3.
- `jenkins.Client.GET(ctx, path, into)` signature in Task 3 matches every call site in
  Tasks 6–11.
- `output.Format` / `output.Detect` / `output.Render` signatures in Task 5 match every call
  site in Tasks 6–11.
- `jenkins.JobPath` in Task 4 is used identically in Tasks 7–9.

---

## Out of scope (deliberately)

- Mutating subcommands (`build`, `abort`, `replay`) — future `--enable-mutations` build tag.
- Pod exec / file ops on the Jenkins controller pod — independent project per earlier
  conversation.
- GitLab / GitHub / Argo Workflows backends — future tasks.
- Distribution (Homebrew tap, Scoop, container image with `ciq` baked in) — covered by
  CI release workflow in a follow-up.
- The kagent runtime container image with `ciq` pre-installed — must be built by the
  user; documented in Task 14's `values.yaml` (`runtimeImage` override).
