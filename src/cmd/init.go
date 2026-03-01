package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init [shell]",
	Short: "Generate shell wrapper script for terminal injection",
	Long: `Generate a shell wrapper script allowing bmh to inject commands directly into the terminal prompt.
Currently supports: zsh, bash, powershell.

Add the following to your shell configuration file:
Zsh (~/.zshrc):   eval "$(bmh init zsh)"
Bash (~/.bashrc): eval "$(bmh init bash)"
PowerShell ($PROFILE): Invoke-Expression (bmh init powershell | Out-String)
`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		shell := args[0]
		switch shell {
		case "zsh":
			fmt.Print(`
function bmh() {
    local tmp_cmd_file=$(mktemp)
    BAOMIHUA_CMD_OUTPUT="$tmp_cmd_file" command bmh "$@"
    
    if [[ -s "$tmp_cmd_file" ]]; then
        local injected_cmd=$(cat "$tmp_cmd_file")
        print -z "$injected_cmd"
    fi
    rm -f "$tmp_cmd_file"
}
alias bmh="bmh"
alias "??"="bmh"
`)
		case "bash":
			fmt.Print(`
function bmh() {
    local tmp_cmd_file=$(mktemp)
    BAOMIHUA_CMD_OUTPUT="$tmp_cmd_file" command bmh "$@"
    
    if [[ -s "$tmp_cmd_file" ]]; then
        local injected_cmd=$(cat "$tmp_cmd_file")
        bind '"\e[0n": "'"$injected_cmd"'"'
        printf '\e[5n'
    fi
    rm -f "$tmp_cmd_file"
}
alias "??"=bmh
`)
		case "powershell":
			exePath, err := os.Executable()
			if err != nil {
				exePath = "bmh.exe"
			}
			fmt.Printf(`
function bmh {
    $tmp_cmd_file = [System.IO.Path]::GetTempFileName()
    $env:BAOMIHUA_CMD_OUTPUT = $tmp_cmd_file
    
    & "%s" @args
    
    if (Test-Path $tmp_cmd_file) {
        $content = Get-Content $tmp_cmd_file -Raw | Out-String
        if (![string]::IsNullOrWhiteSpace($content)) {
            $content = $content.TrimEnd()
            
            $csharp = @"
using System;
using System.Runtime.InteropServices;
public class Keyboard {
    public static void Send(string keys) {
        foreach (char c in keys) {
            string s = c.ToString();
            if ("^+%%~()[]{}".Contains(s)) {
                s = "{" + s + "}";
            }
            System.Windows.Forms.SendKeys.SendWait(s);
        }
    }
}
"@
            if (-not ("Keyboard" -as [type])) {
                Add-Type -TypeDefinition $csharp -ReferencedAssemblies System.Windows.Forms
            }
            [Keyboard]::Send($content)
        }
        Remove-Item $tmp_cmd_file -Force
    }
}
Set-Alias -Name "??" -Value "bmh"
`, exePath)
		default:
			fmt.Printf("Unsupported shell: %s. Supported shells are zsh, bash, powershell.\n", shell)
		}
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
