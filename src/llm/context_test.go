package llm

import (
	"strings"
	"testing"
)

func TestBuildSystemPrompt(t *testing.T) {
	ctx := EnvContext{
		OS:    "windows",
		Shell: "powershell",
		CWD:   "C:\\Users\\test",
	}

	prompt := BuildSystemPrompt(ctx)
	if !strings.Contains(prompt, "windows") {
		t.Errorf("Expected prompt to contain OS info")
	}
	if !strings.Contains(prompt, "powershell") {
		t.Errorf("Expected prompt to contain Shell info")
	}
	if !strings.Contains(prompt, "C:\\Users\\test") {
		t.Errorf("Expected prompt to contain CWD info")
	}
}
