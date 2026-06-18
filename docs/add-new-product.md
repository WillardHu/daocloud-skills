# 如何为新产品添加 CLI 支持

本文档说明 daocloud-skills 的工作原理，以及如何为一个新的 DCE 产品（例如 `mspider`）接入生成式 CLI。

---

## 项目工作原理

整个流程分三层：

```
OpenAPI Spec (JSON)
      ↓  specsync  拉取并缓存到 .cache/
      ↓  codegen   读取 Spec + Overlay 生成 Go 代码和 Skill 文档
      ↓
bin/dce          ← 可执行 CLI，子命令按产品分组
skills/dce/      ← AI Agent Skill（SKILL.md + references/）
```

### 关键文件

| 文件 | 作用 |
|---|---|
| `specs/sources.yaml` | 声明每个产品的 OpenAPI 来源（repo、commit、文件路径、`display_name`） |
| `internal/overlay/<source>.yaml` | 对生成结果做增量修改：隐藏内部 API、补充 short/alias/example/默认分页 |
| `cli.yaml` | CLI 名称、auth 校验接口，以及 `skill.include`（指向手写 Skill 内容目录） |
| `internal/skill-include/` | **手写**的 Skill 散文，codegen 后按相对路径**追加**到生成的 `skills/dce/` 对应文件（CLI 级或模块级，详见下文[手写 Skill 内容放哪](#手写-skill-内容放哪)）。**不要直接改 `skills/dce/`，它每次 codegen 被整体重写** |
| `internal/generated/modules_gen.go` | **自动生成**，把所有产品模块挂载到 cobra root，不要手改 |
| `skills/dce/references/modules/<display_name>.md` | **自动生成**，AI Agent 用的命令索引 |

### 命名：`source key` 与 `display_name`

每个产品在 `sources.yaml` 里有两个名字：

| | **source key** | **display_name** |
|---|---|---|
| 定义位置 | `sources.yaml` 顶层 key | 同一条目下的 `display_name` 字段 |
| 用途 | 内部标识 | CLI 第二段模块名（用户可见） |
| 示例 | `ghippo` | `global-management` |
| 用于 | `make sync-one SOURCE=`、overlay 路径、生成代码目录、缓存目录 | `dce <name> ...` 命令、`skills/.../modules/*.md` 文件名 |

省略 `display_name` 时，CLI 模块名默认等于 source key（如 `insight`）。

**命名约定**：`display_name` 应设置为与 **DCE 控制台页面导航栏** 中对应模块的英文名称一致（通常为 kebab-case）。这样用户在 Web UI 与 CLI 之间能直接对应，例如导航栏「Global Management」→ `global-management`，「Container Management」→ `container-management`。

| DCE 导航栏（英文） | `display_name` | source key |
|---|---|---|
| Global Management | `global-management` | `ghippo` |
| Container Management | `container-management` | `kpanda` |
| Insight | `insight` | `insight` |
| Operations Management | `operations-management` | `gmagpie` |

新增产品时，先在 DCE 控制台确认该模块在导航栏上的英文名称，再填写 `display_name`；source key 仍可使用内部代号（如 OpenAPI 目录名 `mspider`）。

### 手写 Skill 内容放哪

先分清你要加的是「**一个完整的独立 skill**」还是「**给现有 dce skill 补内容**」。

> **判据**：用户说一句话（如"诊断我的集群"），你希不希望它**自己跳出来被触发**？
> - 希望 → 它是**一个完整独立 skill**（自带 `name`/`description`、意图触发），哪怕只服务某个模块的领域（如集群诊断）。走下面的 A。
> - 不希望，只是给 dce skill 补说明 / 注意 / 命令示例 → 走下面的 B。

#### A. 完整独立 skill（含模块级的诊断 / 工作流类）

直接新建 `skills/<name>/SKILL.md`，**手写、直接提交**——它不是生成物，也不走 skill-include / overlay：

- 与 `skills/dce/` **平级**。lathe 只重写 `skills/dce`，**不会碰**你的目录；CI 的 `make bootstrap` + diff 检查也不受影响（bootstrap 只重生成 `skills/dce`）。
- **不要加 `.lathe-skill` 标记**（那是 lathe 生成物的所有权标记，加了反而会被当成可清除目录）。
- 命名建议用冒号命名空间、且对齐 **`display_name`**：如 `container-management:cluster-diagnosis`（而非 source key `kpanda:`），与 dce skill 的模块名一致、description 也更易被触发。
- 装上后靠**自己的 `description` 自发现**，不需要被任何东西引用；若它调用 `dce` 命令，那是运行期依赖（需 dce CLI 一并装好）。
- **不要**把这类塞进 skill-include 或 overlay——那会让它丢掉独立触发能力，退化成一段没人主动召唤的背景文字。

#### B. 给现有 dce skill 补文案（散文 / 按命令的结构化）

`skills/dce/` 整个目录每次 codegen 被 `RemoveAll` 重写，**手改不会保留**，所以通过下面两个通道注入：

| 内容 | 放哪（手写源） | 追加 / 渲染到 |
|---|---|---|
| **CLI 级、跨模块**的散文（如 Module availability） | `internal/skill-include/SKILL.md` | `skills/dce/SKILL.md` |
| **模块级**的散文（绑定某产品、不绑定单条命令，如"GPU 指标约 30s 延迟"） | `internal/skill-include/references/modules/<display_name>.md` | 该模块的 `references/modules/<display_name>.md` |
| **按命令的结构化**（example / note / prerequisite / ignore / short / alias…） | `internal/overlay/<source>.yaml` 的 `commands:` | 渲染进模块文档对应命令段 |

要点：

- **散文走 skill-include，按命令的结构化走 overlay**（overlay 的 schema 只有 `defaults` 和按命令的 `commands:`，放不下模块级散文；但能放某条命令的 `long`/`example`/`notes`/`prerequisites`/`known_errors`）。
- skill-include 是**按相对路径**映射的：`internal/skill-include/<rel>` → `skills/dce/<rel>`。所以模块级散文要用 **`display_name`**（如 `container-management.md`），不是 source key（`kpanda`），否则会生成一个没人读的孤儿文件。
- 追加规则（lathe v0.3.0）：只有 `SKILL.md` 和 `references/**.md` 会**追加**到已有文件；`agents/`、`assets/`、`scripts/` 及 references 下的非-md 文件是**只新建、不追加**（保护生成的非-md 文件）。
- 模块级散文能被 agent 读到，是因为 `SKILL.md` 的 References 列表本就指向 `references/modules/<display_name>.md`。

---

## 添加新产品的完整步骤

以 `mspider`（服务网格，CLI 模块名 `service-mesh`）为例。

### 第一步：在 `specs/sources.yaml` 添加来源

```yaml
sources:
  ghippo:   # 已有；source key = ghippo，CLI 模块名 = global-management
    display_name: global-management
    ...
  kpanda:   # 已有；source key = kpanda，CLI 模块名 = container-management
    display_name: container-management
    ...

  mspider:                                              # 新增；source key
    display_name: service-mesh                          # 与 DCE 导航栏英文模块名一致
    repo_url: https://github.com/DaoCloud/daocloud-api-docs.git
    pinned_tag: <commit-sha>                            # 固定到具体 commit
    backend: swagger
    swagger:
      files:
        - docs/openapi/mspider/v0.x.x.json             # Spec 文件在 repo 中的路径
```

`pinned_tag` 填写 `daocloud-api-docs` 仓库的 commit SHA，确保可复现。`display_name` 优先与 DCE 导航栏英文模块名对齐；若与 source key 相同（如 `insight`），可省略该字段。

### 第二步：同步 Spec 并预览生成结果

```bash
# 只同步 mspider，不影响其他产品
make sync-one SOURCE=mspider
```

执行后：
- Spec 文件缓存到 `.cache/specs-sync/mspider/`（source key）
- `internal/generated/mspider/` 下生成 Go 命令代码（source key）
- `skills/dce/references/modules/service-mesh.md` 生成命令索引（`display_name`）
- `internal/generated/modules_gen.go` 自动追加 mspider 模块

### 第三步：创建 Overlay 文件

```bash
touch internal/overlay/mspider.yaml
```

Overlay 的作用是在不改动生成代码的前提下，控制命令的可见性和质量。

**最小模板：**

```yaml
commands:
  # ── 平台内部 API，不对外暴露 ──────────────────────────
  some-internal-api:   { ignore: true }   # 隐藏该命令

  # ── 服务列表 ──────────────────────────────────────────
  list-services:
    short: "List managed services in the mesh"
    aliases: [ls, ls-svc]
    example: |
      dce service-mesh service list-services --mesh default-mesh
    params:
      page:
        default: "1"
      pageSize:
        default: "20"
```

**Overlay 支持的字段：**

| 字段 | 说明 |
|---|---|
| `ignore: true` | 隐藏该命令，不生成 CLI 子命令 |
| `short` | 命令简介（覆盖 Spec 的 summary） |
| `aliases` | 命令别名列表 |
| `example` | 使用示例，会出现在 `--help` 和 AI Skill 文档中 |
| `params.<name>.default` | 参数默认值（常用于分页） |
| `params.<name>.help` | 参数帮助文字 |

**建议优先处理：**
1. 用 `ignore: true` 屏蔽所有纯内部 API（FeatureGate、Setting、Stream 等）
2. 对高频命令补充 `short`、`aliases`、`example`
3. 所有带分页的命令加上 `page: { default: "1" }` 和 `pageSize: { default: "20" }`

### 第四步：重新生成并验证

```bash
# 重新跑一次 codegen（sync-one 已经跑过 specsync，直接 codegen 即可）
make codegen

# 编译验证
make build

# 检查新产品的子命令是否挂载（用 display_name，不是 source key）
./bin/dce service-mesh --help

# 搜索验证
./bin/dce search "list services" --json
```

### 第五步：构建并推送镜像

```bash
make image-push IMAGE_REPO=release-ci.daocloud.io/clawos/dce-cli IMAGE_TAG=v0.x.x
```

---

## 升级已有产品的 Spec 版本

```bash
# 1. 修改 specs/sources.yaml 中对应产品的 pinned_tag 和 swagger.files 路径
# 2. 重新同步（SOURCE 填 source key，如 ghippo）
make sync-one SOURCE=<source>

# 3. 检查 overlay 是否需要补充新增命令
# 4. 编译验证
make build
```

---

## 常见问题

**Q: codegen 生成的命令太多，怎么批量屏蔽？**

在 overlay 里用 `ignore: true` 逐条屏蔽，或者先不加 overlay 跑一次 `make build`，通过 `./bin/dce <display_name> --help` 看有哪些命令，再决定屏蔽哪些。

**Q: 新产品的 auth 路径和 ghippo 不同怎么办？**

`cli.yaml` 中的 `auth.validate.path` 是全局的登录校验接口，目前固定为 ghippo 的 `/apis/ghippo.io/v1alpha1/current-user`。如果新产品有独立 auth，需要评估是否修改该接口或在 lathe 框架层扩展。

**Q: `make sync-one` 失败，提示无法访问 GitHub？**

网络问题。可以手动把对应的 OpenAPI JSON 文件放到 `.cache/specs-sync/<source>/` 下，并创建 `.cache/specs-sync/<source>/sync-state.yaml`（内容随意），再直接运行 `make codegen`。
