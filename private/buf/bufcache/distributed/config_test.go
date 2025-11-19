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
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfigYAML(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "cache.yaml")

	yamlContent := `
enabled: true
type: s3
bucket: my-bucket
prefix: buf/v1
region: us-east-1
`
	require.NoError(t, os.WriteFile(configPath, []byte(yamlContent), 0644))

	config, err := LoadConfig(configPath)
	require.NoError(t, err)

	assert.True(t, config.Enabled)
	assert.Equal(t, "s3", config.Backend.Type)
	assert.Equal(t, "my-bucket", config.Backend.Bucket)
	assert.Equal(t, "buf/v1", config.Backend.Prefix)
	assert.Equal(t, "us-east-1", config.Backend.Region)
}

func TestLoadConfigJSON(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "cache.json")

	jsonContent := `{
  "enabled": true,
  "backend": {
    "type": "gcs",
    "bucket": "my-gcs-bucket",
    "prefix": "buf/cache"
  }
}`
	require.NoError(t, os.WriteFile(configPath, []byte(jsonContent), 0644))

	config, err := LoadConfig(configPath)
	require.NoError(t, err)

	assert.True(t, config.Enabled)
	assert.Equal(t, "gcs", config.Backend.Type)
	assert.Equal(t, "my-gcs-bucket", config.Backend.Bucket)
}

func TestLoadConfigNonexistent(t *testing.T) {
	t.Parallel()

	config, err := LoadConfig("/nonexistent/path/cache.yaml")
	require.NoError(t, err)
	assert.False(t, config.Enabled)
}

func TestLoadConfigFromEnv(t *testing.T) {
	// Set environment variables
	t.Setenv("BUF_CACHE_DISTRIBUTED_ENABLED", "true")
	t.Setenv("BUF_CACHE_BACKEND_TYPE", "http")
	t.Setenv("BUF_CACHE_BACKEND_URL", "https://cache.example.com")
	t.Setenv("BUF_CACHE_BACKEND_TOKEN_ENV", "MY_TOKEN")

	config := LoadConfigFromEnv()

	assert.True(t, config.Enabled)
	assert.Equal(t, "http", config.Backend.Type)
	assert.Equal(t, "https://cache.example.com", config.Backend.URL)
	assert.Equal(t, "MY_TOKEN", config.Backend.TokenEnv)
}

func TestBackendConfigValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		config  BackendConfig
		wantErr bool
	}{
		{
			name:    "valid s3",
			config:  BackendConfig{Type: "s3", Bucket: "my-bucket"},
			wantErr: false,
		},
		{
			name:    "s3 missing bucket",
			config:  BackendConfig{Type: "s3"},
			wantErr: true,
		},
		{
			name:    "valid gcs",
			config:  BackendConfig{Type: "gcs", Bucket: "my-bucket"},
			wantErr: false,
		},
		{
			name:    "gcs missing bucket",
			config:  BackendConfig{Type: "gcs"},
			wantErr: true,
		},
		{
			name:    "valid http",
			config:  BackendConfig{Type: "http", URL: "https://example.com"},
			wantErr: false,
		},
		{
			name:    "http missing url",
			config:  BackendConfig{Type: "http"},
			wantErr: true,
		},
		{
			name:    "valid local",
			config:  BackendConfig{Type: "local", Path: "/tmp/cache"},
			wantErr: false,
		},
		{
			name:    "local missing path",
			config:  BackendConfig{Type: "local"},
			wantErr: true,
		},
		{
			name:    "unknown type",
			config:  BackendConfig{Type: "unknown"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfigValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name:    "disabled config",
			config:  Config{Enabled: false},
			wantErr: false,
		},
		{
			name: "enabled with valid backend",
			config: Config{
				Enabled: true,
				Backend: BackendConfig{Type: "s3", Bucket: "my-bucket"},
			},
			wantErr: false,
		},
		{
			name: "enabled with invalid backend",
			config: Config{
				Enabled: true,
				Backend: BackendConfig{Type: "s3"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
