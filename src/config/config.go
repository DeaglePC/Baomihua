package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// VendorConfig defines configuration for a specific LLM vendor
type VendorConfig struct {
	Name    string
	APIKey  string
	BaseURL string
}

// AppConfig defines the application configuration
type AppConfig struct {
	Model   string `mapstructure:"model"`
	Vendors []VendorConfig
}

var Cfg AppConfig

// Define default vendor configurations (OpenAI-compatible)
var DefaultVendors = []VendorConfig{
	{Name: "openai", BaseURL: "https://api.openai.com/v1"},
	{Name: "deepseek", BaseURL: "https://api.deepseek.com/v1"},
	{Name: "qwen", BaseURL: "https://dashscope.aliyuncs.com/compatible-mode/v1"},
	{Name: "glm", BaseURL: "https://open.bigmodel.cn/api/paas/v4"},
	{Name: "kimi", BaseURL: "https://api.moonshot.cn/v1"},
	{Name: "minimax", BaseURL: "https://api.minimax.chat/v1"},
	{Name: "claude", BaseURL: "https://api.anthropic.com/v1"}, // May need specific provider logic later
	{Name: "gemini", BaseURL: "https://generativelanguage.googleapis.com/v1beta/openai"},
	{Name: "ernie", BaseURL: "https://aip.baidubce.com/rpc/2.0/ai_custom/v1/wenxinworkshop/chat"}, // Requires custom adapter mapping potentially
}

// InitConfig initializes viper and loads configuration
func InitConfig() {
	// 1. Initial setup for Viper
	viper.SetDefault("model", "gpt-4o")

	// Set config file search paths
	home, err := os.UserHomeDir()
	if err == nil {
		configDir := home + "/.baomihua"
		// Ensure the directory exists
		os.MkdirAll(configDir, 0755)
		viper.AddConfigPath(configDir)
		viper.AddConfigPath(home) // Fallback to home dir for legacy
	}
	viper.AddConfigPath(".")
	viper.SetConfigName("config") // Look for config.yaml inside ~/.baomihua/ or .baomihua.yaml in ~
	viper.SetConfigType("yaml")

	// Read configuration file if exists
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			fmt.Println("Error reading config file:", err)
		}
	}

	// 2. Parse general config
	if err := viper.Unmarshal(&Cfg); err != nil {
		fmt.Println("Error unmarshaling config:", err)
	}

	// Override model via environment variables if present
	if envModel := os.Getenv("BAOMIHUA_MODEL"); envModel != "" {
		Cfg.Model = envModel
	}

	// 3. Dynamically load vendor configurations from Env or Config file
	for _, dv := range DefaultVendors {
		vendorUpper := strings.ToUpper(dv.Name)
		apiKeyEnvName := fmt.Sprintf("%s_API_KEY", vendorUpper)
		baseURLEnvName := fmt.Sprintf("%s_BASE_URL", vendorUpper)

		apiKey := os.Getenv(apiKeyEnvName)
		baseURL := os.Getenv(baseURLEnvName)

		// Fallback to config file if not in env
		if apiKey == "" {
			apiKey = viper.GetString(fmt.Sprintf("%s-api-key", dv.Name))
		}
		if baseURL == "" {
			baseURL = viper.GetString(fmt.Sprintf("%s-base-url", dv.Name))
		}

		// Use default base URL if nothing overridden
		if baseURL == "" {
			baseURL = dv.BaseURL
		}

		if apiKey != "" {
			Cfg.Vendors = append(Cfg.Vendors, VendorConfig{
				Name:    dv.Name,
				APIKey:  apiKey,
				BaseURL: baseURL,
			})
		}
	}

	// 4. Load Custom Vendors generically
	customVendors := viper.GetStringMapString("vendors")
	for name, url := range customVendors {
		// Check if it's already a default vendor
		isDefault := false
		for _, v := range Cfg.Vendors {
			if v.Name == name {
				isDefault = true
				break
			}
		}

		if !isDefault {
			// For custom vendors, API key can be optional (e.g. for local Ollama)
			// But we still try to read it if provided via env: OLLAMA_API_KEY
			apiKeyEnvName := fmt.Sprintf("%s_API_KEY", strings.ToUpper(name))
			apiKey := os.Getenv(apiKeyEnvName)
			if apiKey == "" {
				apiKey = viper.GetString(fmt.Sprintf("%s-api-key", name))
			}

			Cfg.Vendors = append(Cfg.Vendors, VendorConfig{
				Name:    name,
				APIKey:  apiKey, // Might be empty for local models, which is fine
				BaseURL: url,
			})
		}
	}
}

// GetModel returns the configured or flag-overridden model
func GetModel() string {
	// viper bind overrides the default Config struct
	if m := viper.GetString("model"); m != "" && m != "gpt-4o" {
		return m
	}
	return Cfg.Model
}

// GetVendorConfig returns the configuration for a specific vendor
func GetVendorConfig(name string) *VendorConfig {
	for _, v := range Cfg.Vendors {
		if v.Name == name {
			return &v
		}
	}
	return nil
}

// GetAllVendors returns a list of all fully configured vendors
func GetAllVendors() []VendorConfig {
	return Cfg.Vendors
}
