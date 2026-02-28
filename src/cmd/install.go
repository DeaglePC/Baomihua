package cmd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"baomihua/llm"

	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install the bmh shell wrapper into your terminal profile",
	Long:  "Automatically detects your shell and appends the 'bmh init' script to your .zshrc, .bashrc, or PowerShell $PROFILE.",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := llm.GetEnvContext()
		shell := ctx.Shell

		var profilePath string
		var initScript string

		// Normalizing shell name
		if strings.Contains(strings.ToLower(shell), "powershell") || strings.Contains(strings.ToLower(shell), "pwsh") {
			shell = "powershell"

			// Find PowerShell Profile path
			profileCmd := exec.Command("powershell", "-NoProfile", "-Command", "[Console]::Out.Write($PROFILE)")
			var out bytes.Buffer
			profileCmd.Stdout = &out
			if err := profileCmd.Run(); err == nil && out.String() != "" {
				profilePath = strings.TrimSpace(out.String())
			} else {
				// Fallback
				homeDir, _ := os.UserHomeDir()
				profilePath = filepath.Join(homeDir, "Documents", "WindowsPowerShell", "Microsoft.PowerShell_profile.ps1")
			}

			exePath, err := os.Executable()
			if err != nil {
				exePath = "bmh.exe"
			}
			initScript = fmt.Sprintf("\n# BaoMiHua Injection\nInvoke-Expression (& \"%s\" init powershell | Out-String)\n", exePath)

		} else if strings.Contains(strings.ToLower(shell), "zsh") {
			shell = "zsh"
			homeDir, _ := os.UserHomeDir()
			profilePath = filepath.Join(homeDir, ".zshrc")
			initScript = "\n# BaoMiHua Injection\neval \"$(bmh init zsh)\"\n"

		} else if strings.Contains(strings.ToLower(shell), "bash") {
			shell = "bash"
			homeDir, _ := os.UserHomeDir()
			profilePath = filepath.Join(homeDir, ".bashrc")
			initScript = "\n# BaoMiHua Injection\neval \"$(bmh init bash)\"\n"

		} else {
			fmt.Printf("❌ Unsupported shell detected: %s. Please run 'bmh init' manually.\n", shell)
			return
		}

		// Ensure the directory exists (especially for PowerShell profile)
		if dir := filepath.Dir(profilePath); dir != "" {
			os.MkdirAll(dir, 0755)
		}

		// Check if already installed
		content, err := os.ReadFile(profilePath)
		if err == nil && strings.Contains(string(content), "# BaoMiHua Injection") {
			fmt.Printf("✅ bmh is already installed in your %s profile (%s).\n", shell, profilePath)
			return
		}

		// Append to profile
		f, err := os.OpenFile(profilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Printf("❌ Failed to open profile file %s: %v\n", profilePath, err)
			return
		}
		defer f.Close()

		if _, err := f.WriteString(initScript); err != nil {
			fmt.Printf("❌ Failed to write to profile file %s: %v\n", profilePath, err)
			return
		}

		fmt.Printf("🎉 Successfully installed bmh wrapper to %s!\n", profilePath)
		fmt.Printf("🔄 Please restart your terminal or run this command to apply changes immediately:\n")

		if shell == "powershell" {
			fmt.Printf("   . \"%s\"\n", profilePath)
		} else {
			fmt.Printf("   source \"%s\"\n", profilePath)
		}
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
}
