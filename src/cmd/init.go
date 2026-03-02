package cmd

import (
	"fmt"
	"os"
	"strings"
)

// InitWrapper generates a shell wrapper script allowing bmh to inject commands
// directly into the terminal prompt. Currently supports: zsh, bash, powershell.
// Add the following to your shell configuration file:
// Zsh (~/.zshrc):   eval "$(bmh --init zsh)"
// Bash (~/.bashrc): eval "$(bmh --init bash)"
// PowerShell ($PROFILE): Invoke-Expression (& bmh --init powershell | Out-String)
func InitWrapper(shell string) {
	switch shell {
	case "zsh":
		fmt.Print(`
function bmh() {
    local tmp_cmd_file=$(mktemp)
    if [[ "$*" == "fuck" ]]; then
        export BAOMIHUA_LAST_CMD=$(fc -ln -1 2>/dev/null || echo "")
    fi
    BAOMIHUA_CMD_OUTPUT="$tmp_cmd_file" command bmh "$@"
    
    if [[ "$*" == "fuck" ]]; then
        unset BAOMIHUA_LAST_CMD
    fi
    
    if [[ -s "$tmp_cmd_file" ]]; then
        local injected_cmd=$(cat "$tmp_cmd_file")
        print -z "$injected_cmd"
    fi
    rm -f "$tmp_cmd_file"
}
alias bmh="bmh"
alias "??"="bmh"
alias fuck="bmh fuck"
`)
	case "bash":
		fmt.Print(`
function bmh() {
    local tmp_cmd_file=$(mktemp)
    if [[ "$*" == "fuck" ]]; then
        export BAOMIHUA_LAST_CMD=$(fc -ln -1 2>/dev/null || echo "")
    fi
    BAOMIHUA_CMD_OUTPUT="$tmp_cmd_file" command bmh "$@"
    
    if [[ "$*" == "fuck" ]]; then
        unset BAOMIHUA_LAST_CMD
    fi
    
    if [[ -s "$tmp_cmd_file" ]]; then
        local injected_cmd=$(cat "$tmp_cmd_file")
        bind '"\e[0n": "'"$injected_cmd"'"'
        printf '\e[5n'
    fi
    rm -f "$tmp_cmd_file"
}
alias "??"=bmh
alias fuck="bmh fuck"
`)
	case "powershell":
		exePath, err := os.Executable()
		if err != nil {
			exePath = "bmh.exe"
		}

		script := `function bmh {
    param([parameter(ValueFromRemainingArguments=$true)] $Rest)

    $isFuck = ($Rest -join " ") -eq "fuck"
    if ($isFuck) {
        if ($global:Error.Count -gt 0) {
            $env:BAOMIHUA_LAST_ERROR = $global:Error[0].Exception.Message
            if ($global:Error[0].InvocationInfo) {
                $env:BAOMIHUA_LAST_CMD = $global:Error[0].InvocationInfo.Line
            } else {
                $env:BAOMIHUA_LAST_CMD = ""
            }
        }
    }

    $tmp_cmd_file = [System.IO.Path]::GetTempFileName()
    $env:BAOMIHUA_CMD_OUTPUT = $tmp_cmd_file
    
    & "{{BMH_EXE}}" $Rest
    
    if ($isFuck) {
        Remove-Item Env:\BAOMIHUA_LAST_ERROR -ErrorAction SilentlyContinue
        Remove-Item Env:\BAOMIHUA_LAST_CMD -ErrorAction SilentlyContinue
    }
    
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
            if ("^+%~()[]{}".Contains(s)) {
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
function fuck { bmh fuck }
`
		fmt.Print(strings.Replace(script, "{{BMH_EXE}}", exePath, 1))
	default:
		fmt.Printf("Unsupported shell: %s. Supported shells are zsh, bash, powershell.\n", shell)
	}
}
