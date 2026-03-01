<h1 align="center">🐆 BaoMiHua</h1>

> **A "Bengal cat" lurking in your terminal—your exclusive, ultra-fast AI command assistant.**

BaoMiHua (`bmh` or `bao`) is a terminal-based AI assistant built with Go. It boasts a blazing-fast cold start experience and an elegant, high-value interactive UI. It sharply perceives your current operating system and shell environment, turning your natural language into precise shell commands while offering a safe, seamless interactive execution experience.

## ✨ Core Features

- ⚡️ **Ultra-fast Cold Start**: Built with Go, natively compiled for instant response—zero waiting time.
- 🧠 **Natural Language to Commands**: Just tell it what you want to do, and it will output the most accurate shell command for you.
- 🕵️ **Intelligent Context Awareness**: Silently collects OS (Windows/macOS/Linux), shell environment (bash/zsh/powershell, etc.), and working directory info, ensuring generated commands are 100% tailored to your current environment.
- 🛡️ **Safety Guard & Interception**: Built-in dangerous command scanner (e.g., `rm -rf /`). When the AI hallucinates or generates a high-risk command, it triggers a highlighted red UI warning and forcefully downgrades operation privileges to prevent catastrophes.
- 🎨 **Elegant Aesthetics**: Features a sleek terminal UI powered by `Bubble Tea`, complete with silky loading animations (`bubbles/spinner`) that breathe life into the cold terminal.
- 🧩 **1-Click Seamless Execution**: Allows you to directly copy, execute, or seamlessly inject the generated command straight into your current terminal prompt.

## 📦 Installation

### One-line Install Script (Recommended)

The fastest way to install. This script automatically detects your system type, downloads the pre-compiled binary, and configures it in your environment variables.

**macOS / Linux**
```bash
curl -fsSL https://raw.githubusercontent.com/DeaglePC/Baomihua/main/install.sh | bash
```

**Windows (PowerShell)**
```powershell
Invoke-RestMethod -Uri https://raw.githubusercontent.com/DeaglePC/Baomihua/main/install.ps1 | Invoke-Expression
```

### Build from Source (Alternative)
Please ensure [Go](https://go.dev/) (1.20+ recommended) is installed on your system.

```bash
git clone https://github.com/DeaglePC/Baomihua.git
cd Baomihua/src
go build -o bmh
# Move bmh to a directory in your system PATH, for example:
# mv bmh /usr/local/bin/
```

If you are on Windows, you can add the directory containing the compiled `bmh.exe` to your system's `Path` environment variable.

## ⚙️ Configuration Guide

BaoMiHua supports flexible multi-platform configuration using a strict priority strategy: **Command-line Arguments > Environment Variables > Local Config File > Default Values**.

### Supported Models

BaoMiHua natively integrates and supports both major international and domestic LLM APIs. The system automatically scans for available models based on configured API Keys. Below is a list of supported vendors and their corresponding API Key environment variables:

| Vendor | Environment Variable | `config.yaml` Key |
| --- | --- | --- |
| **OpenAI / ChatGPT** | `OPENAI_API_KEY` | `openai-api-key` |
| **DeepSeek** | `DEEPSEEK_API_KEY` | `deepseek-api-key` |
| **Qwen** | `QWEN_API_KEY` | `qwen-api-key` |
| **GLM (Zhipu)** | `GLM_API_KEY` | `glm-api-key` |
| **Kimi (Moonshot)** | `KIMI_API_KEY` | `kimi-api-key` |
| **MiniMax** | `MINIMAX_API_KEY` | `minimax-api-key` |
| **Claude (Anthropic)** | `CLAUDE_API_KEY` | `claude-api-key` |
| **Gemini (Google)** | `GEMINI_API_KEY` | `gemini-api-key` |
| **Ernie (Baidu)** | `ERNIE_API_KEY` | `ernie-api-key` |

### Configuring Vendors & API Keys

BaoMiHua allows configuring API Keys independently across different **Model Vendors**. You can configure keys for OpenAI, Gemini, Claude, and various Chinese vendors simultaneously, and the tool will manage them collectively.

When listing (`--list`) or switching models, **the system will only display available models from vendors for which you have configured an API Key**. 
To ensure model lists stay up-to-date, the specific lists are not hardcoded but are **dynamically fetched from each vendor's API and cached locally**.

| Configuration Item | Env Var Example | Config File (`~/.baomihua/config.yaml`) Example | Description |
| --- | --- | --- | --- |
| **Vendor API Key** | `OPENAI_API_KEY`<br>`DEEPSEEK_API_KEY`<br>... | `openai-api-key: "sk-..."`<br>`deepseek-api-key: "sk-..."` | Format: `{VENDOR}_API_KEY`. Once configured, models from this vendor will be fetched, unlocked, and displayed. |
| **Global Default Model** | `BAOMIHUA_MODEL` | `model: "gpt-4o"` | Global default model. Can be temporarily overridden using `bmh --model=xxx`. |
| **Vendor Interface URL**| `OPENAI_BASE_URL` | `openai-base-url: "..."` | Optional. Used for proxies, self-hosted proxy APIs, etc. (configured per vendor). |
| **Custom Vendor (e.g., Ollama)**| N/A | `vendors:`<br>&nbsp;&nbsp;`ollama: "http://127.0.0.1:11434/v1"` | Connect to any local or private API compatible with the OpenAI `/v1/chat/completions` standard. The dictionary key is used as the vendor name, and the value is the Base URL. The system will look for a `{VendorName}_API_KEY` env var automatically. |

#### Full Configuration Example: `~/.baomihua/config.yaml`

You can reference the snippet below for comprehensive configuration inside `~/.baomihua/config.yaml`:

```yaml
# Global default model (override via --model flag)
model: deepseek-coder

# Native vendor API Key configs (Env vars have higher priority)
deepseek-api-key: "sk-xxxxxxxxxxxxxxxxxxxxxxxx"
openai-api-key: "sk-proj-yyyyyyyyyyyyyyyyyyyyyyyy"
qwen-api-key: "sk-zzzzzzzzzzzzzzzzzzzzzzzz"

# Custom / Local deployed vendors (OpenAI API standard compatible)
vendors:
  # Local Ollama connection
  ollama: "http://127.0.0.1:11434/v1"
  
  # Another private company deployed LLM endpoint
  mycorp: "http://192.168.1.100:8080/v1"
```

*Example: Configuring multiple vendors simultaneously (Windows PowerShell)*
```powershell
$env:OPENAI_API_KEY="sk-xxxxxxxxxxx"
$env:DEEPSEEK_API_KEY="sk-yyyyyyyyyyy"
$env:BAOMIHUA_MODEL="deepseek-coder"
```

*Example: Configuring multiple vendors simultaneously (Linux/macOS)*
```bash
export OPENAI_API_KEY="sk-xxxxxxxxxxx"
export DEEPSEEK_API_KEY="sk-yyyyyyyyyyy"
export BAOMIHUA_MODEL="deepseek-coder"
```

## 🚀 Quick Start

Extremely simple usage: just ask it what you would normally ask Google or an LLM chatbot to get commands for:

```bash
bmh Find the process occupying port 8080 and forcefully kill it
```

**Interactive Menu Options:**
After analyzing the model's response, the breakdown will be displayed, and you will be presented with the following interactive menu:
1. 🐾 **Insert to prompt**: Injects the command right at your shell input cursor. Just hit Enter to execute. *(Recommended default)*
2. ⚡️ **Execute**: Immediately run the command and throw the output directly back to the terminal. *(Disabled entirely if a high-risk command is detected)*
3. 📋 **Copy**: Copies the generated command into your system clipboard.
4. 🛑 **Cancel**: Exit the current dialogue flow.

## 🛠️ Tech Stack & Tooling

- Routing / CLI Framework: [Cobra](https://github.com/spf13/cobra)
- Configuration Management: [Viper](https://github.com/spf13/viper)
- Terminal UI & State Machine: [Bubble Tea](https://github.com/charmbracelet/bubbletea)
- CSS Styling / Engine: [Lip Gloss](https://github.com/charmbracelet/lipgloss)

---
> 🐆 *BaoMiHua - A fierce beast in the terminal, taming the complex CLI world for you.*
