# ciq — AI-agent-friendly CI inspection CLI

Read-only CLI for querying CI/CD systems (Jenkins first; GitLab CI / GitHub Actions planned).
Designed to be safe for AI agents to invoke: structured JSON output, compile-time read-only guarantee,
auth tokens never leak into command output or LLM context.

Status: alpha, Jenkins read-only commands.

(Full README in Task 13.)
