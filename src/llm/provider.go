package llm

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"baomihua/config"
)

// Provider interface defines how a vendor is interacted with.
type Provider interface {
	Name() string
	GetAvailableModels() ([]string, error)
	StreamCompletion(model, prompt string, ctx EnvContext, contentChan chan<- string, errChan chan<- error)
}

// ModelRegistry holds the active providers and cached models
type ModelRegistry struct {
	providers []Provider
	models    map[string]string // maps model name -> vendor name
	mu        sync.RWMutex
}

var GlobalRegistry *ModelRegistry

func InitRegistry() {
	GlobalRegistry = &ModelRegistry{
		models: make(map[string]string),
	}

	vendors := config.GetAllVendors()
	for _, v := range vendors {
		// Currently all supported listed vendors can use the OpenAI compatible REST API
		GlobalRegistry.providers = append(GlobalRegistry.providers, NewOpenAICompatibleProvider(v))
	}
}

// LoadModels attempts to load models from cache, or fetches them via APIs
func (r *ModelRegistry) LoadModels(forceRefresh bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	cacheFile := getCachePath()
	if !forceRefresh {
		if cached, err := loadModelsFromCache(cacheFile); err == nil && len(cached) > 0 {
			r.models = cached
			return nil
		}
	}

	// Fetch concurrently
	var wg sync.WaitGroup
	var mapMu sync.Mutex
	newModels := make(map[string]string)

	for _, p := range r.providers {
		wg.Add(1)
		go func(prov Provider) {
			defer wg.Done()
			models, err := prov.GetAvailableModels()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to fetch models from %s: %v\n", prov.Name(), err)
				return
			}
			mapMu.Lock()
			for _, m := range models {
				newModels[m] = prov.Name()
			}
			mapMu.Unlock()
		}(p)
	}

	wg.Wait()
	r.models = newModels

	// Save to cache
	saveModelsToCache(cacheFile, newModels)
	return nil
}

func (r *ModelRegistry) GetModelsList() map[string][]string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	list := make(map[string][]string)
	for m, v := range r.models {
		list[v] = append(list[v], m)
	}
	return list
}

func (r *ModelRegistry) GetProviderForModel(model string) (Provider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	vendorName, ok := r.models[model]
	if !ok {
		// Fallback routing heuristics if model not cached but provided explicitly via flag
		for _, p := range r.providers {
			if strings.Contains(strings.ToLower(model), p.Name()) {
				return p, nil
			}
		}
		// If only 1 provider exists, fallback to it
		if len(r.providers) == 1 {
			return r.providers[0], nil
		}
		return nil, fmt.Errorf("model '%s' not found in cache and could not determine vendor. Run 'bmh --list' to refresh or check your API keys", model)
	}

	for _, p := range r.providers {
		if p.Name() == vendorName {
			return p, nil
		}
	}
	return nil, fmt.Errorf("provider '%s' for model '%s' not configured", vendorName, model)
}

func (r *ModelRegistry) GetProviderByName(name string) (Provider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, p := range r.providers {
		if strings.EqualFold(p.Name(), name) {
			return p, nil
		}
	}
	return nil, fmt.Errorf("provider '%s' not configured", name)
}

// Cache structs
type modelsCache struct {
	Timestamp time.Time         `json:"timestamp"`
	Models    map[string]string `json:"models"` // model -> vendor
}

func getCachePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".baomihua_models.json"
	}
	configDir := filepath.Join(home, ".baomihua")
	os.MkdirAll(configDir, 0755)
	return filepath.Join(configDir, "models.json")
}

func loadModelsFromCache(path string) (map[string]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cache modelsCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, err
	}

	// Expire cache after 24 hours
	if time.Since(cache.Timestamp).Hours() > 24 {
		return nil, fmt.Errorf("cache expired")
	}

	return cache.Models, nil
}

func saveModelsToCache(path string, models map[string]string) {
	cache := modelsCache{
		Timestamp: time.Now(),
		Models:    models,
	}
	data, err := json.Marshal(cache)
	if err == nil {
		_ = os.WriteFile(path, data, 0644)
	}
}

// --- Specific Provider Implementations ---

// OpenAICompatibleProvider handles generic OpenAI format endpoints used by many vendors
type OpenAICompatibleProvider struct {
	vendor config.VendorConfig
}

func NewOpenAICompatibleProvider(v config.VendorConfig) *OpenAICompatibleProvider {
	return &OpenAICompatibleProvider{vendor: v}
}

func (p *OpenAICompatibleProvider) Name() string {
	return p.vendor.Name
}

type openAIModelsResponse struct {
	Data []struct {
		ID string `json:"id"`
	} `json:"data"`
}

func (p *OpenAICompatibleProvider) GetAvailableModels() ([]string, error) {
	if p.vendor.Name == "minimax" {
		// Minimax API does not natively support the /v1/models endpoint to list available models.
		// Return a static list of their most common models.
		return []string{
			"MiniMax-M2.5",
			"MiniMax-M2.5-highspeed",
			"MiniMax-M2.1",
			"MiniMax-M2.1-highspeed",
			"MiniMax-M2",
		}, nil
	}

	url := strings.TrimRight(p.vendor.BaseURL, "/") + "/models"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+p.vendor.APIKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var res openAIModelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}

	var models []string
	for _, m := range res.Data {
		models = append(models, m.ID)
	}
	return models, nil
}
