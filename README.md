# 🐆 BaoMiHua (豹米花)

> **一个存在于终端里的“孟加拉豹猫”——你的专属极速 AI 终端指令助手。**

BaoMiHua (`bmh` 或 `bao`) 是一个基于 Go 语言构建的终端 AI 助手。它拥有极速的冷启动体验和高颜值的交互界面 (UI)。它能敏锐感知当前的操作系统与 Shell 环境，将你的自然语言精妙地转化为精准的 Shell 命令，并提供安全、无缝的交互执行体验。

## ✨ 核心特性

- ⚡️ **极速冷启动**：采用 Go 语言构建，原生编译，拒绝等待，即刻响应。
- 🧠 **自然语言转命令**：只需告诉它你想做什么，它会为你输出最准确的 Shell 指令。
- 🕵️ **智能上下文感知**：静默收集 OS (Windows/macOS/Linux)、Shell 环境 (bash/zsh/powershell 等) 及工作目录信息，让生成的指令 100% 契合当前环境。
- 🛡️ **安全防御与拦截 (Safety Guard)**：内置危险命令扫描器（例如 `rm -rf /`）。当 AI 产生幻觉或生成高危指令时，触发 UI 红色高亮警告，并强制降级操作权限，防患于未然。
- 🎨 **高颜值交互**：基于 `Bubble Tea` 提供优雅的终端 UI，丝滑的加载动画 (`bubbles/spinner`)，让冰冷的终端也充满灵动。
- 🧩 **一键无缝执行**：支持将生成的命令直接复制、执行，或利用 Shell 特性无缝插入到当前终端 prompt 中。

## 📦 安装

### 一键安装脚本 (推荐)

最快速的安装方式，这会自动检测你的系统类型并下载安装预编译好的二进制文件并配置到环境变量。

**macOS / Linux**
```bash
curl -fsSL https://raw.githubusercontent.com/DeaglePC/Baomihua/main/install.sh | bash
```

**Windows (PowerShell)**
```powershell
Invoke-RestMethod -Uri https://raw.githubusercontent.com/DeaglePC/Baomihua/main/install.ps1 | Invoke-Expression
```

### 源码编译安装 (备选)
请确保系统已安装 [Go](https://go.dev/) (推荐 1.20 及以上版本)。

```bash
git clone https://github.com/DeaglePC/Baomihua.git
cd Baomihua/src
go build -o bmh
# 将 bmh 移动到系统的 PATH 目录下，例如：
# mv bmh /usr/local/bin/
```

若是 Windows 环境，可以将编译出的 `bmh.exe` 目录添加到系统的环境变量 Path 中。

## ⚙️ 配置指南

BaoMiHua 支持灵活的多端配置，采用严格的优先级策略：**命令行参数 > 环境变量 > 本地配置文件 > 默认值**。

### 支持的模型

BaoMiHua （豹米花）原生集成并支持国内外主流的大模型 API。系统会自动根据配置的 Key 扫描对应的可用模型。以下是目前代码中已适配的厂商列表及对应的 API Key 环境变量名：

| 厂商 (Vendor) | 环境变量名 (Environment Variable) | `config.yaml` 键名 |
| --- | --- | --- |
| **OpenAI / ChatGPT** | `OPENAI_API_KEY` | `openai-api-key` |
| **DeepSeek (深度求索)** | `DEEPSEEK_API_KEY` | `deepseek-api-key` |
| **Qwen (通义千问)** | `QWEN_API_KEY` | `qwen-api-key` |
| **GLM (智谱清言)** | `GLM_API_KEY` | `glm-api-key` |
| **Kimi (月之暗面)** | `KIMI_API_KEY` | `kimi-api-key` |
| **MiniMax** | `MINIMAX_API_KEY` | `minimax-api-key` |
| **Claude (Anthropic)** | `CLAUDE_API_KEY` | `claude-api-key` |
| **Gemini (Google)** | `GEMINI_API_KEY` | `gemini-api-key` |
| **Ernie (文心一言)** | `ERNIE_API_KEY` | `ernie-api-key` |

### 厂商与 API Key 配置

BaoMiHua 支持按不同的**模型厂商（Vendor）**分别独立配置 API Key。您可以同时配置 OpenAI、Gemini、Claude 以及国内诸多厂商的 Key，工具会自动统筹管理。

在列出（`--list`）或切换模型时，**系统只会展示您已经配置了 API Key 的厂商旗下的可用模型**。
为保证模型列表的时效性，具体的模型列表并非固化在代码中，而是**通过调用各厂商的 API 动态拉取并缓存在本地**的，确保您总能第一时间使用到最新的模型。

| 配置项 | 环境变量格式示例 | 配置文件 (`~/.baomihua/config.yaml`) 示例 | 说明 |
| --- | --- | --- | --- |
| **厂商 API 密钥** | `OPENAI_API_KEY`<br>`DEEPSEEK_API_KEY`<br>... | `openai-api-key: "sk-..."`<br>`deepseek-api-key: "sk-..."` | 格式为 `{VENDOR}_API_KEY`。配置后对应厂商的模型才会被拉取并解锁展示。 |
| **全局默认模型** | `BAOMIHUA_MODEL` | `model: "gpt-4o"` | 全局默认使用的模型。也支持经由 `bmh --model=xxx` 临时覆盖。 |
| **厂商接口地址**| `OPENAI_BASE_URL` | `openai-base-url: "..."` | 可选。用于支持代理、自建中转 API 等（按厂商独立配置）。 |
| **自定义厂商 (如 Ollama)**| 无 (纯配置) | `vendors:`<br>&nbsp;&nbsp;`ollama: "http://127.0.0.1:11434/v1"` | 如果你需要接入任何兼容 OpenAI `/v1/chat/completions` 标准的其他本地或私有 API，可以在配置文件中用 `vendors` 属性字典来自定义。字典的 Key 会作为厂商名称，Value 则是 Base URL。系统会自动给这个厂商寻找 `{厂商名}_API_KEY` 的环境变量（如果有的话）。 |

#### `~/.baomihua/config.yaml` 完整配置样例

你可以参考以下样例，在 `~/.baomihua/config.yaml` 中进行全量配置：

```yaml
# 全局默认使用的模型 (可通过 --model 参数覆盖)
model: deepseek-coder

# 原生支持的厂商 API Key 配置 (环境变量优先级更高，这里作为补充或替代)
deepseek-api-key: "sk-xxxxxxxxxxxxxxxxxxxxxxxx"
openai-api-key: "sk-proj-yyyyyyyyyyyyyyyyyyyyyyyy"
qwen-api-key: "sk-zzzzzzzzzzzzzzzzzzzzzzzz"

# 自定义私有 / 本地部署厂商 (兼容 OpenAI 标准)
vendors:
  # 接入本地的 Ollama
  ollama: "http://127.0.0.1:11434/v1"
  
  # 接入局域网的其他私有化部署大模型
  mycorp: "http://192.168.1.100:8080/v1"
```

*示例：同时配置多个厂商 (Windows PowerShell)*
```powershell
$env:OPENAI_API_KEY="sk-xxxxxxxxxxx"
$env:DEEPSEEK_API_KEY="sk-yyyyyyyyyyy"
$env:BAOMIHUA_MODEL="deepseek-coder"
```

*示例：同时配置多个厂商 (Linux/macOS)*
```bash
export OPENAI_API_KEY="sk-xxxxxxxxxxx"
export DEEPSEEK_API_KEY="sk-yyyyyyyyyyy"
export BAOMIHUA_MODEL="deepseek-coder"
```

## 🚀 快速开始

极其简单的操作方式，把你原本需要 Google 或去询问大模型的命令问题告诉它即可：

```bash
bmh 查出占用 8080 端口的进程并强制杀掉
```

**交互选项：**
模型分析完毕后，会展示解析详情，并提供以下交互菜单选项：
1. 🐾 **插入终端 (Insert to prompt)**：将命令放入输入框光标处，由您确认后敲击回车。*(默认推荐)*
2. ⚡️ **直接执行 (Execute)**：即刻运行，并将结果直接抛回终端展示。*(若检测为高危命令，将禁用此选项)*
3. 📋 **复制命令 (Copy)**：将生成的命令送入系统剪贴板。
4. 🛑 **放弃 (Cancel)**：退出当前对话。

## 🛠️ 技术栈选型

- 路由基建：[Cobra](https://github.com/spf13/cobra)
- 配置管理：[Viper](https://github.com/spf13/viper)
- 终端界面与状态机：[Bubble Tea](https://github.com/charmbracelet/bubbletea)
- CSS 样式渲染引擎：[Lip Gloss](https://github.com/charmbracelet/lipgloss)

---
> 🐆 *BaoMiHua - 终端里的一只猛兽，为你驯服繁杂的命令行世界。*