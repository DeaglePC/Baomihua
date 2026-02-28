package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

// UpdateDefaultModel writes the new model to ~/.baomihua/config.yaml
func UpdateDefaultModel(newModel string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	configPath := filepath.Join(home, ".baomihua", "config.yaml")

	data, err := os.ReadFile(configPath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	content := string(data)
	if content == "" {
		content = fmt.Sprintf("model: %s\n", newModel)
	} else {
		re := regexp.MustCompile(`(?m)^model:\s*.*$`)
		if re.MatchString(content) {
			content = re.ReplaceAllString(content, fmt.Sprintf("model: %s", newModel))
		} else {
			content = fmt.Sprintf("model: %s\n%s", newModel, content)
		}
	}

	return os.WriteFile(configPath, []byte(content), 0644)
}
