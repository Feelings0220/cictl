# Contributing

- Conventional Commits required: `feat:`, `fix:`, `docs:`, `test:`, `chore:`, `refactor:`.
- Sign off commits with `git commit -s`.
- Every new Jenkins endpoint method needs an `httptest`-backed table-driven test.
- No mutating HTTP verbs in `internal/jenkins/client.go` without an RFC-style design issue first.
