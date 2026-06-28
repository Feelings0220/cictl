# 在 kagent 里把 `cictl` 跑起来

kagent 的内置 Agent runtime 镜像 (`ghcr.io/kagent-dev/kagent/app:<ver>`) 默认不带
`cictl`。要在 kagent Agent 里调 `cictl jenkins ...`，你需要让二进制出现在 Agent
容器的 `PATH` 里。下面两种方案，按真实场景挑。

---

## 方案 A：自定义 runtime 镜像（生产推荐）

把 `cictl` 烧进 kagent runtime 镜像，发到自己的镜像仓库，然后让 Agent 的
deployment 用这个镜像。

### 1. 构建镜像

```bash
docker build \
  --build-arg KAGENT_VERSION=v0.x.y \
  --build-arg CICTL_VERSION=v0.1.0 \
  -t ghcr.io/<your-namespace>/kagent-runtime-with-cictl:v0.1.0 \
  -f docker/Dockerfile.runtime .

docker push ghcr.io/<your-namespace>/kagent-runtime-with-cictl:v0.1.0
```

`docker/Dockerfile.runtime` 是个 multi-stage build：第 1 阶段在 alpine
里从源码编译 cictl（静态链接，CGO 关闭），第 2 阶段 `FROM` kagent 的官方
runtime 镜像并 COPY 二进制进去。

> **重要**：`KAGENT_VERSION` 必须跟你部署的 kagent 控制面版本对齐——不一致可能
> 导致 controller 调度 Pod 时镜像里的 ADK 协议跟 controller 期待的对不上。

### 2. 让 Agent CRD 用你的镜像

kagent 的 Helm chart 顶层 values 暴露了 `app.agentImage.{registry, repository, tag}`：

```yaml
# kagent values.yaml
app:
  agentImage:
    registry: ghcr.io
    repository: <your-namespace>/kagent-runtime-with-cictl
    tag: v0.1.0
```

这会让 **所有** Declarative Agent 都用这个镜像——如果你不想全局生效，单 Agent
覆盖见 kagent 文档的 `Agent.spec.declarative.deployment.imageRegistry`。

---

## 方案 B：sidecar + emptyDir（实验、临时）

不重新打镜像，靠 sidecar 把 cictl 放进共享 volume，主容器 PATH 里加一条。

**限制**：kagent v1alpha2 的 `SharedDeploymentSpec` **没有** `initContainers`
字段——只有 `extraContainers`（sidecar）。所以 cictl 的可用窗口要等 sidecar
启动完成、写完文件之后；主容器要在 PATH 里加上共享目录。两点都要你自己保证。

不推荐，但如果你不能修改全局镜像，这是个 workaround。具体接法见 kagent 的
`agent.deploymentSpec` helper 源码。

---

## 方案 C：MCP server（未来路线）

把 cictl 的能力包成一个真正的 MCP server，注册为 `RemoteMCPServer` CRD，
Agent 通过 `tools: [{type: McpServer, ...}]` 引用。这是 kagent 内置 Agent
（k8s、istio、observability）采用的模式。**目前不是 cictl 的范围**——见
项目 README 的"路线图"章节。

---

## 凭据

`cictl` 读 `~/.config/cictl/credentials.yaml`。kagent Agent 容器里 `HOME`
通常是 `/home/agent`。挂载 K8s Secret 到那里：

```yaml
# Agent.spec.declarative.deployment 片段
volumes:
  - name: cictl-credentials
    secret:
      secretName: jenkins-cictl-credentials
      items:
        - key: credentials.yaml
          path: credentials.yaml
volumeMounts:
  - name: cictl-credentials
    mountPath: /home/agent/.config/cictl
    readOnly: true
env:
  - name: HOME
    value: /home/agent
```

对应的 Secret：

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: jenkins-cictl-credentials
  namespace: kagent
stringData:
  credentials.yaml: |
    default-context: prod
    contexts:
      prod:
        url: https://jenkins.example.com
        username: alice
        token: <jenkins-api-token>
```
