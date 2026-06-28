# kagent PR Checklist

目标：把 `helm/agents/jenkins-triage/` 加上一个 Agent CRD 样例落到 kagent 仓库，
作为一个 PR 提交。

## 提交前检查

1. **核对 CRD 字段。** 打开 kagent 仓库的 `go/api/v1alpha2/agent_types.go`，确认：
   - `Agent.spec.skills.gitRefs[*]` 存在且接受 `url` / `ref` / `path` 三个字段。
   - `Agent.spec.runtime` 暴露 `image` / `env` / `volumeMounts`。
   - `Agent.spec.volumes` 接受 `[]corev1.Volume` 形状的切片。
   如果字段名不同，相应调整 `templates/agent.yaml`。
2. **匹配现有 chart 风格。** 对比 `helm/agents/observability/`：
   - `Chart-template.yaml` 的 key
   - `_helpers.tpl` 的命名约定
   - `values.yaml` 的 key 名和缩进
3. **本地 lint。** 在 kagent 仓库根目录运行 `helm lint helm/agents/jenkins-triage`。

## Vendoring 到 kagent

在 cictl 仓库根目录，把 `KAGENT_DIR` 指向你 fork 的 kagent 路径：

```bash
make vendor-to-kagent KAGENT_DIR=$HOME/code/kagent
```

这一步会把 `kagent-pr/helm-agents-jenkins-triage/` 复制到
`$KAGENT_DIR/helm/agents/jenkins-triage/`。

## 提交 + 开 PR

在 kagent fork 里：

```bash
git checkout -b feat/jenkins-triage-agent
git add helm/agents/jenkins-triage
git commit -s -m "feat: add jenkins-triage agent (read-only via cictl)"
git push origin feat/jenkins-triage-agent
```

PR 描述模板：

```markdown
## What
新增一个内置 agent `jenkins-triage`，用于只读地诊断 Jenkins 构建失败。

## How
所有 Jenkins 查询都通过开源 CLI `cictl`（https://github.com/Feelings0220/cictl）。
通过 `Agent.spec.skills.gitRefs` 加载 `jenkins` skill。结构上是只读的——
`cictl` 这一构建里没有任何 mutation 子命令。

## Test plan
- [ ] `helm lint helm/agents/jenkins-triage`
- [ ] 在 Kind 集群部署，对接一个可达的 Jenkins，创建
      `jenkins-cictl-credentials` Secret，验证 agent 能正确回答
      "列出 team/service folder 下的任务"。
- [ ] 验证排查过程中 Jenkins access log 里没有任何 POST / PUT / DELETE 请求。
```

## 关于运行时镜像

`cictl` 二进制必须能被 Agent runtime 调用。两种方式：

1. **自定义 runtime 镜像（推荐）。** 写一个 Dockerfile，从 kagent 官方 runtime 镜像
   FROM 起，把 `cictl` 二进制 COPY 进去。Push 到自己的镜像仓库，在 `values.yaml` 里
   把 `runtimeImage` 改成它。

2. **initContainer 下载（实验场景）。** 在 Pod 里加一个 initContainer 从 GitHub
   Releases 下载 `cictl`，挂到 emptyDir，主容器 `PATH` 加上这个 emptyDir。简单但每次
   起 Pod 都会拉一次网络。

生产环境务必走方案 1。
