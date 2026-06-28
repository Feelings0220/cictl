# 架构

## 分层

```
┌──────────────────┐
│  cmd/ciq         │  cobra 子命令；不直接接触 HTTP
│  （每个资源一个   │
│   文件）         │
└────────┬─────────┘
         │ 调用
         ▼
┌──────────────────┐
│ internal/jenkins │  每个资源一个文件（jobs / builds / console / clouds / queue）
│                  │  每个方法都有 httptest 支撑的 table-driven 测试
└────────┬─────────┘
         │ 使用
         ▼
┌──────────────────┐
│ internal/jenkins │  client.go：HTTP client 结构体，仅 GET，basic auth，
│  client.go       │           错误中的 token 会被清洗
└────────┬─────────┘
         │ 使用
         ▼
┌──────────────────┐
│ internal/config  │  基于 YAML 的凭据加载器，支持多 context 切换
└──────────────────┘

┌──────────────────┐
│ internal/output  │  json / table / md 格式化器，自动检测 TTY
└──────────────────┘
```

## 只读是结构性的，不是开关

`internal/jenkins/client.go` 只导出一个方法：`GET(ctx, path, into)`。
**没有** `POST` / `PUT` / `DELETE` / `PATCH` 辅助方法可以被滥用。
要在未来加入 mutation 能力，必须修改这个文件——而这正是 review 阶段
最容易拦下的位置。

CI 流水线里加一道防线就更稳：

```bash
! grep -E 'http.MethodPost|http.MethodPut|http.MethodDelete|http.MethodPatch' internal/jenkins/client.go
```

只要这一行 grep 命中，构建就该失败。

## Token 清洗

Go 的 `net/url` 错误会把 URL 嵌进 `error.Error()` 字符串里，如果 URL 里带有
凭据（通常不会，但可能发生），token 就会泄露。`client.go` 里的 `scrubURLError`
显式把 token 字符串替换成 `***`，防止它进入日志或 LLM 上下文。

`internal/config/config.go` 在 YAML 解析失败时**不**回显文件内容——
文件可能包含 token，所以只返回 `parse credentials yaml: invalid syntax`，
不带任何原始字节。

## 添加新的 CI 后端（GitLab CI / GitHub Actions）

1. 在 `internal/<name>/` 下新建一个包，镜像 `internal/jenkins/` 的结构：
   `client.go`、`path.go`（如果需要）、按资源拆分的文件。
2. 在 `cmd/ciq/<name>.go` 添加父命令，按资源拆分子命令文件。
3. 每加一个资源类型，先写 `httptest` 支撑的测试，再暴露 CLI。
4. 写一份 `skills/<name>.md` 描述对 Agent 暴露的命令面。

`cmd/ciq/root.go` 里的 `rootFlags`、`internal/output`、`internal/config`
都是后端无关的，新后端可以直接复用。

## 文件职责一览

| 路径 | 职责 |
|---|---|
| `cmd/ciq/main.go` | 入口，调用 `newRoot().Execute()` |
| `cmd/ciq/root.go` | cobra root command，全局 flag |
| `cmd/ciq/jenkins.go` | `ciq jenkins` 父命令 + 子命令注册 + `newJenkinsClient` 辅助 |
| `cmd/ciq/jenkins_*.go` | 各资源的 CLI 子命令（一文件一资源） |
| `internal/config/config.go` | 加载 `credentials.yaml`，按 context 解析 |
| `internal/jenkins/client.go` | HTTP client（GET only） |
| `internal/jenkins/path.go` | `team/service/main` → `/job/team/job/service/job/main` |
| `internal/jenkins/<resource>.go` | 各资源的 endpoint 方法（list / get / 等） |
| `internal/output/formatter.go` | JSON / table / md 渲染器 + TTY 检测 |
