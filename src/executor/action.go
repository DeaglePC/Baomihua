package executor

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"baomihua/llm"

	"github.com/atotto/clipboard"
)

// ExecuteCommand runs a command in the current OS shell directly
func ExecuteCommand(cmdStr string, ctx llm.EnvContext) error {
	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		if strings.Contains(strings.ToLower(ctx.Shell), "powershell") || strings.Contains(strings.ToLower(ctx.Shell), "pwsh") {
			cmd = exec.Command("powershell", "-NoProfile", "-Command", cmdStr)
		} else {
			cmd = exec.Command("cmd", "/c", cmdStr)
		}
	} else {
		cmd = exec.Command("sh", "-c", cmdStr)
	}

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// CopyToClipboard copies the text to the system clipboard
func CopyToClipboard(text string) error {
	return clipboard.WriteAll(text)
}

// InjectToTerminal writes the command to the file specified by BAOMIHUA_CMD_OUTPUT env var
func InjectToTerminal(cmdStr string) error {
	outputFile := os.Getenv("BAOMIHUA_CMD_OUTPUT")
	if outputFile == "" {
		return fmt.Errorf("BAOMIHUA_CMD_OUTPUT environment variable is not set. Ensure you run this via the wrapper script (try running 'bmh init')")
	}

	return os.WriteFile(outputFile, []byte(cmdStr), 0600)
}
