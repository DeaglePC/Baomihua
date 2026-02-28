package guard

import (
	"regexp"
	"strings"
)

// Level represents the safety level of a command
type Level int

const (
	Normal Level = iota
	Danger
)

// dangerousPatterns is a list of regular expressions that match potentially dangerous commands
var dangerousPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)\brm\s+(?:.*?\s+)?-[A-Za-z\s-]*r[A-Za-z\s-]*(?:/|/\*|~)(?:\s+|$)`), // Recursive force remove root/home
	regexp.MustCompile(`(?i)\bmkfs\b`),                                                                // Format filesystem                // Format filesystem
	regexp.MustCompile(`(?i)\bdd\s+.*of=/dev/`),                                                       // Destructive dd to block devices
	regexp.MustCompile(`(?i)\bchmod\s+-R\s+777\s+/(?:\s+|$)`),                                         // Recursive chmod 777 on root
	regexp.MustCompile(`(?i)>\s*/dev/(?:sd[a-z]+|hd[a-z]+|vd[a-z]+|xvd[a-z]+|nvme[0-9]+|disk[0-9]+)`), // Overwrite block devices directly
}

// CheckCommand evaluates a shell command and determines its safety level
func CheckCommand(command string) Level {
	cmdStr := strings.TrimSpace(command)
	if cmdStr == "" {
		return Normal
	}

	for _, pattern := range dangerousPatterns {
		if pattern.MatchString(cmdStr) {
			return Danger
		}
	}

	return Normal
}
