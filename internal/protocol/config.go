package protocol

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type UIConfig struct {
	Theme           string `json:"theme"`
	ViewMode        string `json:"viewMode"`
	SortMode        string `json:"sortMode"`
	BannerCollapsed bool   `json:"bannerCollapsed"`
}

func DefaultConfig() UIConfig {
	return UIConfig{
		Theme:    "cyan",
		ViewMode: "flat",
		SortMode: "recent",
	}
}

func ConfigPath() string {
	return filepath.Join(RegistryDir(), "config.json")
}

func ReadConfig() UIConfig {
	cfg := DefaultConfig()
	data, err := os.ReadFile(ConfigPath())
	if err != nil {
		return cfg
	}
	json.Unmarshal(data, &cfg)
	return cfg
}

func WriteConfig(cfg UIConfig) {
	os.MkdirAll(RegistryDir(), 0755)
	data, _ := json.MarshalIndent(cfg, "", "  ")
	os.WriteFile(ConfigPath(), append(data, '\n'), 0644)
}
