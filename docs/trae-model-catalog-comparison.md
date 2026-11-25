# Trae 模型目录版本标识对比报告

## 1. 结论

已使用提供的 Trae 账户信息，直接调用上游模型目录接口，复现了 cli2api v0.7.45 的 Trae 模型缺失问题。

结论如下：

- 同一账号、同一接口、同一请求体，只切换 Trae 版本标识头。
- 旧标识，即 **cli2api v0.7.44 当时使用的 Trae 标识**，返回完整模型目录。
- 新标识，即 **cli2api v0.7.45 引入的 Trae 标识**，返回明显缩小的模型目录。
- 新标识没有带回任何旧标识没有的模型。
- 缺失内容与用户反馈的“只缺新模型、全部入口都缺”一致。
- 根因可以确认：`f9639d4` 修改 Trae 客户端标识后，上游按新版本标识返回了更旧的 catalog。


---

## 2. 对比条件

### 接口

```text
POST https://trae-api-cn.mchost.guru/api/ide/v1/get_detail_param
```

### 请求体

分别测试两个 catalog scene。

`chat_v3`：

```json
{
  "function": "chat_v3",
  "config_names": null,
  "need_prompt": false,
  "current_config_info": null,
  "poly_prompt": true,
  "mode_type": null,
  "agent_type": null
}
```

`solo_work_lite`：

```json
{
  "function": "solo_work_lite",
  "config_names": null,
  "need_prompt": false,
  "current_config_info": null,
  "poly_prompt": true,
  "mode_type": null,
  "agent_type": null
}
```

### 版本标识组

| 分组 | 对应版本 | `X-Ide-Version` / `User-Agent` | `X-Ide-Version-Code` / `X-App-Version-Code` |
|---|---|---:|---:|
| 旧标识 | cli2api v0.7.44 使用的 Trae 标识 | `3.3.62` | `20260811` |
| 新标识 | cli2api v0.7.45 引入的 Trae 标识 | `3.4.24` | `20260212` |

---

## 3. 总体结果

HTTP 状态码全部为 `200`，没有接口错误。

| Scene | 旧标识模型数 | 新标识模型数 | 仅旧标识有 | 仅新标识有 |
|---|---:|---:|---:|---:|
| `chat_v3` | 65 | 45 | 20 | 0 |
| `solo_work_lite` | 54 | 15 | 39 | 0 |

关键现象：

- 新标识返回数量显著减少。
- `only new` 全为空，说明新版本标识没有任何增益。
- `only old` 中包含大量新模型，说明旧标识才是当前完整 catalog。

---

## 4. `chat_v3` 差异

旧标识返回 65 个模型，新标识返回 45 个模型。

### 仅旧标识返回的 20 个模型

```text
Doubao-Seed-2.1-Pro
Doubao-Seed-2.1-Turbo
Doubao-Seed-Evolving
deepseek-v4.1-flash
glm-5.3
glm-5.3-flash
glm-5.3-flashx
glm-5.4
kimi-k2.7-code
kimi-k2.8-preview
kimi-k3
kimi-k3.1-pro
kimi-k3.1-pro-highspeed
mimo-v2.6-flash
mimo-v2.6-pro
minimax-m3.1-flash
qwen3.8-flash
qwen3.8-max
qwen4-max
step-5-preview
```

### 仅新标识返回的模型

```text
无
```

---

## 5. `solo_work_lite` 差异

旧标识返回 54 个模型，新标识返回 15 个模型。

### 仅旧标识返回的 39 个模型

```text
DeepSeek-V4-Flash
DeepSeek-V4-Flash-Official
DeepSeek-V4-Pro
DeepSeek-V4-Pro-Official
Doubao-Seed-2.0-Code
Doubao-Seed-2.1-Pro
Doubao-Seed-2.1-Turbo
Doubao-Seed-Evolving
aquila
browser_use_subagent
computer_use_subagent
custom_model_doubao_1M
custom_model_doubao_256k
custom_model_gpt-6
custom_model_tokenpony_deepseek_v4_1_flash
custom_model_tokenpony_glm_5_3
custom_model_tokenpony_qwen3_8_max
deepseek-v4.1-flash
explore_sub_agent_v2
glm-5.2
glm-5.3
glm-5.3-flash
glm-5.3-flashx
glm-5.4
kimi-k2.6
kimi-k2.7-code
kimi-k3
kimi-k3.1-pro
kimi-k3.1-pro-highspeed
mimo-v2.6-flash
mimo-v2.6-pro
minimax-m3
minimax-m3.1-flash
qwen-3.7-plus
qwen3.8-max
qwen4-max
sagitta
seed-code-pro-0430
step-5-preview
```

### 仅新标识返回的模型

```text
无
```

---

## 6. 根因分析

cli2api v0.7.45 中的提交：

```text
f9639d4 feat(providers): align client identities with current official builds (#587)
```

把 Trae 标识从：

```go
IdeVersion     = "3.3.62"
IdeVersionCode = "20260811"
PluginVersion  = "2.3.62834"
```

改成了：

```go
IdeVersion     = "3.4.24"
IdeVersionCode = "20260212"
PluginVersion  = "2.3.106768"
```

模型目录请求实际使用的是：

```text
User-Agent: Trae/<IdeVersion>
X-Ide-Version: <IdeVersion>
X-Ide-Version-Code: <IdeVersionCode>
X-App-Version-Code: <IdeVersionCode>
```

直接 curl 已证明：上游 `get_detail_param` 会根据这些版本标识返回不同 catalog。

因此：

- 这不是控制台过滤问题。
- 不是 gateway 聚合问题。
- 不是 provider policy 过滤问题。
- 是上游 catalog 请求在 v0.7.45 用了新 Trae 标识后，返回了旧目录。

这也能解释为什么用户看到的是“全部入口都只缺新模型”：所有入口最终都读取同一个上游 catalog。

---

## 7. 已实施修复

### `internal/providers/trae/credential.go`

恢复 cli2api v0.7.44 当时使用的 Trae 标识：

```go
IdeVersion     = "3.3.62"
IdeVersionCode = "20260811"
PluginVersion  = "2.3.62834"
```

并明确注释这是：

```text
the Trae client identity values used by cli2api v0.7.44
```

避免再次被误解为 Trae 客户端版本。

### `internal/providers/trae/headers_test.go`

新增回归测试：

```go
TestSetCatalogHeadersUseKnownGoodBuildIdentity
```

断言模型目录请求头为：

```text
User-Agent: Trae/3.3.62
X-Ide-Version: 3.3.62
X-Ide-Version-Code: 20260811
X-App-Version-Code: 20260811
```

测试注释中包含本次实测数据：

```text
chat_v3 45 vs 65, solo 15 vs 54
```

### `changelog/unreleased/fix-trae-model-catalog.md`

新增中英文 changelog fragment，说明恢复 cli2api v0.7.44 使用的 Trae 客户端标识，让模型目录请求重新包含新发布模型。

---

## 8. 验证结果

```text
go test ./internal/providers/trae/...                                      ok
go test ./internal/control/... ./internal/gateway/... ./internal/console/... ok
python3 scripts/release-notes.py validate                                  ok
git diff --check                                                           ok
```

当前工作区状态：

```text
 M internal/providers/trae/credential.go
 M internal/providers/trae/headers_test.go
?? changelog/unreleased/fix-trae-model-catalog.md
```

---

## 9. 最终判断

这次回归已经通过真实账号 curl 验证：

> cli2api v0.7.45 把 Trae 版本标识改成了 `3.4.24 / 20260212`，导致上游模型目录接口返回旧 catalog；恢复 cli2api v0.7.44 使用的 Trae 标识 `3.3.62 / 20260811` 后，新模型恢复。

当前修复是最低风险方案：只回退导致 catalog 缩减的 Trae 标识，不改动 v0.7.45 的其他功能。
