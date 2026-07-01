---
name: jenkins
description: Query Jenkins read-only (jobs, builds, console, clouds, queue) via the cictl CLI. Safe to use during incident triage; never mutates state.
---

# Jenkins (read-only) via `cictl`

You have access to a CLI named `cictl` that wraps the Jenkins REST API. **All operations are GET-only — this binary cannot trigger, abort, or modify any Jenkins object.** Treat this as a safe inspection tool.

## Prerequisites

1. **`cictl` on `PATH`.** Check with `cictl --version`. If it is missing:
   - Run the bundled installer: `bash "$CLAUDE_PLUGIN_ROOT/skills/jenkins/scripts/install-cictl.sh"` (detects OS/arch, downloads the matching release from GitHub, verifies the checksum).
   - Or `go install github.com/Feelings0220/cictl/cmd/cictl@latest`.
   - Or download a binary from https://github.com/Feelings0220/cictl/releases and put it on `PATH`.
   - In kagent, `cictl` is baked into the Agent runtime image (see the repo's `docker/Dockerfile.runtime`); no install step is needed at runtime.
2. **Credentials.** `cictl` reads `~/.config/cictl/credentials.yaml`. If it is missing, copy `examples/credentials.yaml.example` from the repo and fill in the Jenkins URL + a read-only API token. Never ask the user to paste a token into the chat.

## When to use

- Diagnose a failed build (read console log, inspect build metadata, fetch job config).
- Inventory jobs across folders.
- Inspect Kubernetes / EC2 cloud configurations (e.g., investigate why pod-templated agents are not scheduling).
- Inspect the build queue.

## Authentication

Credentials live in `~/.config/cictl/credentials.yaml` (see Prerequisites). If multiple environments are available, switch with `--context <name>` (e.g. `--context staging`). Never ask the user for credentials interactively.

## Output

Add `--format json` for structured output you can parse. Without that flag the output is a human table; if stdout is not a terminal, `cictl` defaults to JSON automatically.

## Command catalog

| What you need | Command |
|---|---|
| Confirm auth works | `cictl jenkins whoami` |
| List jobs (top-level) | `cictl jenkins job list` |
| List jobs in a folder | `cictl jenkins job list --folder team/service` |
| Job metadata + last build | `cictl jenkins job get <name>` |
| Full job config.xml (pipeline def, params, triggers) | `cictl jenkins job config <name>` |
| List recent builds | `cictl jenkins build list <name> --limit 10` |
| One build's metadata (result, duration, cause) | `cictl jenkins build get <name> <number>` |
| Last 200 lines of console log | `cictl jenkins console <name> <number>` |
| Full console log | `cictl jenkins console <name> <number> --full` |
| Last 50 lines (smaller context) | `cictl jenkins console <name> <number> --tail 50` |
| List configured clouds | `cictl jenkins cloud list` |
| One cloud's config.xml (e.g. k8s pod templates) | `cictl jenkins cloud get <name>` |
| What's queued right now | `cictl jenkins queue list` |

## Job names with folders

Jenkins folders nest jobs. Pass the full path with slashes:

```
cictl jenkins job get org/team/service-main
cictl jenkins console org/team/service-main 42
```

## Triage playbook (build failed)

1. `cictl jenkins build get <job> <num> --format json` — get `result`, `duration`, `url`.
2. `cictl jenkins console <job> <num> --tail 300` — usually enough to spot the failure line.
3. If the failure mentions a missing agent / pod, run `cictl jenkins cloud list` and `cictl jenkins cloud get <cloud>` to inspect Kubernetes pod templates and limits.
4. If the failure is config-shaped (wrong branch, missing parameter), run `cictl jenkins job config <job>` to view the pipeline definition.
5. If many builds are queued, `cictl jenkins queue list` and look for `stuck:true` items.

## Hard rules

- **Never** suggest a `cictl` command that would mutate Jenkins. The CLI does not support it. If the user explicitly asks to retry/abort/replay, tell them to use the Jenkins UI or their `gh`/`glab`/`kubectl` equivalents — do not invoke `curl` to Jenkins yourself.
- **Never** print or echo the credentials file. The user can manage it.
- Console logs can be huge — default to `--tail 200` and only escalate to `--full` if you've located the failure region.
