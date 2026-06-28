# ciq —— 给 AI Agent 用的 CI 巡查 CLI

`ciq` 是一个小巧的 Go CLI，让 AI Agent（以及人类）能够安全地查询 CI/CD 系统。
Jenkins 是第一个后端；GitLab CI 和 GitHub Actions 在路线图上。

## 为什么需要它

Agent 直接用 `curl` 调 Jenkins 既脆弱（CSRF crumb、folder 路径、XML 解析），
又不安全（它可以构造 POST 触发或中止任务）。把 API 包到一个专用 CLI 里，能获得三个收益：

1. **编译期只读保证。** 本构建的 HTTP client 没有 `POST` / `PUT` / `DELETE` 辅助方法。
   任何子命令都无法改动 Jenkins 状态。
2. **LLM 友好的输出。** 每个命令都支持 `--format json|table|md`；输出是结构化、稳定的。
3. **集中式认证。** Token 放在 `~/.config/ciq/credentials.yaml`，永远不会出现在命令输出、
   错误消息或 LLM 上下文里。

## 安装

```
go install github.com/kagent-dev/ciq/cmd/ciq@latest
```

预编译二进制：见 Releases。

## 配置

把 `examples/credentials.yaml.example` 复制到 `~/.config/ciq/credentials.yaml`，
填入 Jenkins API Token（生成路径：**People → 你的用户名 → Configure → API Token**）。

```yaml
default-context: prod
contexts:
  prod:
    url: https://jenkins.prod.example.com
    username: alice
    token: REPLACE_ME
  staging:
    url: https://jenkins.staging.example.com
    username: alice
    token: REPLACE_ME
    insecure: true  # 仅用于自签证书的实验环境
```

## 使用

```
ciq jenkins whoami                                  # 验证认证是否通了
ciq jenkins job list                                # 列出顶层任务
ciq jenkins job list --folder team/service          # 列出 folder 下的任务
ciq jenkins job get team/service/main               # 任务元数据 + 最近构建
ciq jenkins job config team/service/main            # 完整 config.xml
ciq jenkins build list team/service/main --limit 10 # 最近 10 次构建
ciq jenkins build get team/service/main 42          # 某次构建的元数据
ciq jenkins console team/service/main 42            # 最后 200 行日志
ciq jenkins console team/service/main 42 --full     # 完整日志
ciq jenkins cloud list                              # 列出云配置（K8s / EC2 等）
ciq jenkins cloud get kubernetes                    # 某个 cloud 的 config.xml
ciq jenkins queue list                              # 当前构建队列
```

多环境？加 `--context staging`。要管道处理？加 `--format json`，输出会变成
结构化 JSON，直接喂给 `jq`。

## 与 AI Agent 配合

把 `skills/jenkins.md` 跟二进制一起发布。任何支持加载 skill markdown 的 Agent 框架——
kagent（`Agent.spec.skills.gitRefs`）、Claude Code（`~/.claude/skills/`）、Cursor——
都可以直接挂上去，让 Agent 开始排查 Jenkins 故障。

## 设计原则

- **只读是结构性的，不是开关。** `internal/jenkins/client.go` 只对外暴露一个方法
  `GET(ctx, path, into)`。没有 `POST` helper 可以被误用。要加 mutation，得改这个文件——
  也就是 review 阶段就能拦住的地方。
- **Token 永远不进入 LLM 上下文。** 配置文件由 CLI 读取，错误消息会做敏感词清洗。
- **LLM 优先的输出。** 所有 read 命令都支持 `--format json`；非 TTY 自动切 JSON。
- **路径就是文档。** 文件结构清晰到从命令到 HTTP endpoint 一眼可见。

## 路线图

- GitLab CI 后端（`ciq gitlab ...`）
- GitHub Actions 后端（`ciq github ...`）
- Argo Workflows、Tekton
- 通过 `--enable-mutations` 编译 tag 引入可选的 mutation 命令（独立 binary）

## License

Apache 2.0.
