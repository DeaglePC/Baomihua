package llm

import (
	"fmt"
	"os"
	"runtime"
)

// EnvContext holds the information about the current terminal environment
type EnvContext struct {
	OS    string
	Shell string
	CWD   string
}

// GetEnvContext collects the current environment variables and OS info
func GetEnvContext() EnvContext {
	shell := os.Getenv("SHELL")
	if shell == "" {
		// Heuristic to detect PowerShell on Windows instead of falling back to cmd.exe immediately
		if os.Getenv("PSModulePath") != "" {
			shell = "powershell (Windows)"
		} else {
			shell = os.Getenv("COMSPEC")
		}
	}
	if shell == "" {
		shell = "unknown"
	}

	cwd, err := os.Getwd()
	if err != nil {
		cwd = "unknown"
	}

	return EnvContext{
		OS:    runtime.GOOS,
		Shell: shell,
		CWD:   cwd,
	}
}

// BuildSystemPrompt generates the system prompt injecting the environment context
func BuildSystemPrompt(ctx EnvContext) string {
	return fmt.Sprintf(`You are a terminal AI assistant named "BaoMiHua" (or "bmh" / "bao").
Your task is to interpret the user's natural language request and provide a precise shell command that safely accomplishes their goal.

CURRENT ENVIRONMENT:
- Operating System: %s
- Shell: %s
- Current Working Directory (CWD): %s

REQUIREMENTS:
1. The generated shell command MUST be compatible with the current OS and Shell.
2. If the user's request is ambiguous or inherently dangerous, output a safe alternative or explain why it cannot be done directly.
3. You MUST return the result in strictly JSON format.
4. Your output MUST be ONLY a JSON object with two string fields:
   - "explanation": A brief, clear explanation of what the command does.
   - "command": The exact shell command to execute.

DO NOT output any markdown (like backticks) around the JSON. ONLY output valid JSON string.
Example JSON output:
{"explanation": "Find the process listening on port 8080 and kill it", "command": "lsof -ti:8080 | xargs kill -9"}
`, ctx.OS, ctx.Shell, ctx.CWD)
}
