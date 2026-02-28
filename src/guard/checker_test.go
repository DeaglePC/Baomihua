package guard

import (
	"testing"
)

func TestCheckCommand(t *testing.T) {
	tests := []struct {
		name     string
		command  string
		expected Level
	}{
		{"Normal echo", "echo 'hello'", Normal},
		{"Normal ls", "ls -l /", Normal},
		{"Danger rm root", "rm -rf /", Danger},
		{"Danger rm root spaces", "rm  -r  -f   /", Danger},
		{"Danger rm root uppercase", "RM -RF /", Danger},
		{"Danger rm root star", "rm -rf /*", Danger},
		{"Danger rm home", "rm -rf ~", Danger},
		{"Normal rm file", "rm -rf ./folder", Normal},
		{"Danger chmod", "chmod -R 777 /", Danger},
		{"Danger mkfs", "mkfs.ext4 /dev/sda1", Danger},
		{"Danger dd", "dd if=/dev/zero of=/dev/sda", Danger},
		{"Danger echo to device", "echo 'hi' > /dev/sda", Danger},
		{"Normal echo to null", "echo 'hi' > /dev/null", Normal},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CheckCommand(tt.command)
			if result != tt.expected {
				t.Errorf("CheckCommand(%q) = %v, want %v", tt.command, result, tt.expected)
			}
		})
	}
}
