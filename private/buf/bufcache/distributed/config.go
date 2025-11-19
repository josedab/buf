// Copyright 2020-2025 Buf Technologies, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package distributed

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the distributed cache configuration.
type Config struct {
	// Enabled indicates whether distributed caching is enabled.
	Enabled bool `json:"enabled" yaml:"enabled"`
	// Backend contains the backend configuration.
	Backend BackendConfig `json:"backend" yaml:"backend,inline"`
}

// DefaultConfig returns a default configuration with distributed cache disabled.
func DefaultConfig() Config {
	return Config{
		Enabled: false,
	}
}

// LoadConfig loads the distributed cache configuration from the given path.
// The path can be a YAML or JSON file.
func LoadConfig(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}
		return Config{}, err
	}

	var config Config
	ext := filepath.Ext(path)
	switch ext {
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &config); err != nil {
			return Config{}, err
		}
	case ".json":
		if err := json.Unmarshal(data, &config); err != nil {
			return Config{}, err
		}
	default:
		// Try YAML first, then JSON
		if err := yaml.Unmarshal(data, &config); err != nil {
			if err := json.Unmarshal(data, &config); err != nil {
				return Config{}, errors.New("failed to parse config as YAML or JSON")
			}
		}
	}

	return config, nil
}

// LoadConfigFromEnv loads the distributed cache configuration from environment variables.
func LoadConfigFromEnv() Config {
	config := DefaultConfig()

	if os.Getenv("BUF_CACHE_DISTRIBUTED_ENABLED") == "true" {
		config.Enabled = true
	}

	if backendType := os.Getenv("BUF_CACHE_BACKEND_TYPE"); backendType != "" {
		config.Backend.Type = backendType
	}

	if bucket := os.Getenv("BUF_CACHE_BACKEND_BUCKET"); bucket != "" {
		config.Backend.Bucket = bucket
	}

	if prefix := os.Getenv("BUF_CACHE_BACKEND_PREFIX"); prefix != "" {
		config.Backend.Prefix = prefix
	}

	if region := os.Getenv("BUF_CACHE_BACKEND_REGION"); region != "" {
		config.Backend.Region = region
	}

	if url := os.Getenv("BUF_CACHE_BACKEND_URL"); url != "" {
		config.Backend.URL = url
	}

	if tokenEnv := os.Getenv("BUF_CACHE_BACKEND_TOKEN_ENV"); tokenEnv != "" {
		config.Backend.TokenEnv = tokenEnv
	}

	if path := os.Getenv("BUF_CACHE_BACKEND_PATH"); path != "" {
		config.Backend.Path = path
	}

	if timeout := os.Getenv("BUF_CACHE_BACKEND_TIMEOUT"); timeout != "" {
		if d, err := time.ParseDuration(timeout); err == nil {
			config.Backend.Timeout = d
		}
	}

	return config
}

// Validate validates the configuration.
func (c Config) Validate() error {
	if !c.Enabled {
		return nil
	}
	return c.Backend.Validate()
}

// GetConfigPath returns the default configuration file path for distributed cache.
func GetConfigPath() string {
	// Check for XDG_CONFIG_HOME first
	if xdgConfigHome := os.Getenv("XDG_CONFIG_HOME"); xdgConfigHome != "" {
		return filepath.Join(xdgConfigHome, "buf", "cache.yaml")
	}

	// Fall back to HOME
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	return filepath.Join(home, ".config", "buf", "cache.yaml")
}

// CacheConfig represents the full cache configuration as it might appear in buf.yaml.
type CacheConfig struct {
	Distributed Config `json:"distributed" yaml:"distributed"`
}
