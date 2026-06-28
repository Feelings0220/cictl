# kagent PR Checklist（B 方案：examples/）

目标：把 `examples/jenkins-triage/README.md` 落到 kagent 仓库，提一个 PR。

**为什么不直接做 `helm/agents/jenkins-triage/` built-in agent？**

读了 kagent v1alpha2 CRD 之后发现：

- 内置 agent（k8s、istio、observability...）全都用 MCP server 当工具，没有
  一个用"CLI + skill markdown"这条路。
- `spec.declarative.deployment` 不能覆盖 main container 镜像，也没有
  `initContainers` 字段——把 `cictl` 二进制放进 Agent 容器里需要用户**自己** build
  一个自定义 runtime 镜像。
- 做成 built-in agent 会让 maintainer 问"为什么不写 MCP server？"，对开局不利。

所以走 examples/ 路径——把 jenkins-triage 当作"如何接入外部 CI CLI"的**文档化范例**，
比建立第 9 个内置 agent 更匹配 kagent 当下的模式。

## 提交前检查

1. **跑通本地端到端测试。**
   ```bash
   make create-kind-cluster   # 在 kagent 仓库根目录
   make helm-install
   # build 自定义 runtime 镜像（见 cictl/docs/runtime-setup.md）
   # apply 你测试用的 Jenkins Secret + Agent CRD
   # 端口转发 UI，跑一遍 prompt
   ```

2. **确认 PR 不动 v1alpha1。** kagent CLAUDE.md 明确规定 v1alpha1 只接受 critical bug fix。
   我们的范例用的全是 v1alpha2。

3. **拼写、Markdown 格式。** 在 kagent 根目录跑 `markdownlint` 如果项目里有配（多数没配）。

## Vendoring 到 kagent fork

在 cictl 仓库根目录：

```bash
make vendor-to-kagent KAGENT_DIR=$HOME/code/kagent
```

这会把 `kagent-pr/examples-jenkins-triage/` 复制到
`$KAGENT_DIR/examples/jenkins-triage/`。**只是一个 README 文件**——比 Helm chart
轻量得多，review 起来快。

## 提交 + 开 PR

在 kagent fork 里（`origin` 已经指向 Feelings0220/kagent，`upstream` 指向
kagent-dev/kagent）：

```bash
git fetch upstream
git checkout -b feat/examples-jenkins-triage upstream/main
git add examples/jenkins-triage
git commit -s -m "docs(examples): add jenkins-triage agent walkthrough (cictl integration)"
git push origin feat/examples-jenkins-triage

gh pr create --repo kagent-dev/kagent \
  --base main \
  --head Feelings0220:feat/examples-jenkins-triage \
  --title "docs(examples): jenkins-triage agent walkthrough (cictl integration)" \
  --body "$(cat <<'EOF'
## What

Adds `examples/jenkins-triage/README.md`, a worked example showing how to wire
an external read-only CI inspection CLI (cictl, https://github.com/Feelings0220/cictl)
into a kagent Agent for Jenkins build-failure triage.

## Why

kagent's existing built-in agents all use MCP servers as tool providers, which is
the right pattern for first-class kagent tools. But there's currently no
documented path for the "I already have a CLI; how do I get my Agent to use it?"
case. This example fills that gap with:

- A skill loaded via `Agent.spec.skills.gitRefs`
- A custom runtime image pattern (Dockerfile lives in the cictl repo)
- A credentials Secret mounted via \`deployment.volumes\` / \`volumeMounts\`
- Inline YAMLs the user can copy-paste

The example explicitly compares its approach against an MCP-server-based
alternative, so readers understand when to choose each.

## Scope

- **One new file**: \`examples/jenkins-triage/README.md\`.
- **No CRD changes.**
- **No new helm chart.** This is documentation only.

## Test plan

- [ ] \`markdownlint examples/jenkins-triage/README.md\` (if project lints docs)
- [ ] Manual walkthrough on a Kind cluster: build the custom runtime image,
      create the Secret, apply the Agent CRD, verify a triage prompt
      surfaces a real failure analysis.
- [ ] Confirm Jenkins access logs show only \`GET\` requests during a triage
      session (read-only guarantee holds end-to-end).

## Related

- cictl repo: https://github.com/Feelings0220/cictl
- runtime setup guide: https://github.com/Feelings0220/cictl/blob/v0.1.0/docs/runtime-setup.md
EOF
)"
```

## 失败回退

如果 maintainer 觉得 examples/ 不该放这种"指向外部工具"的内容，备选：

1. 改成发到 **kagent discussion**（GitHub Discussions），先聊设计共识。
2. 把同样的内容贴到 cictl repo 的 README，加一行 "想接 kagent 看这里"，
   不动 kagent 仓库。
