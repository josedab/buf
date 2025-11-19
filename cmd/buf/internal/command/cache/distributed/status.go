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
	"context"
	"fmt"

	"buf.build/go/app/appcmd"
	"buf.build/go/app/appext"
	"github.com/bufbuild/buf/private/buf/bufcache/distributed"
	"github.com/spf13/pflag"
)

const (
	configFlagName = "config"
)

func newStatusCommand(
	name string,
	builder appext.SubCommandBuilder,
) *appcmd.Command {
	flags := newStatusFlags()
	return &appcmd.Command{
		Use:   name,
		Short: "Show distributed cache status",
		Long:  "Display the current status and statistics of the distributed build cache.",
		Run: builder.NewRunFunc(
			func(ctx context.Context, container appext.Container) error {
				return runStatus(ctx, container, flags)
			},
		),
		BindFlags: flags.Bind,
	}
}

type statusFlags struct {
	Config string
}

func newStatusFlags() *statusFlags {
	return &statusFlags{}
}

func (f *statusFlags) Bind(flagSet *pflag.FlagSet) {
	flagSet.StringVar(
		&f.Config,
		configFlagName,
		"",
		"Path to cache configuration file",
	)
}

func runStatus(
	ctx context.Context,
	container appext.Container,
	flags *statusFlags,
) error {
	// Load configuration
	var config distributed.Config
	var err error
	if flags.Config != "" {
		config, err = distributed.LoadConfig(flags.Config)
	} else {
		configPath := distributed.GetConfigPath()
		config, err = distributed.LoadConfig(configPath)
		if err != nil {
			// Fall back to environment variables
			config = distributed.LoadConfigFromEnv()
			err = nil
		}
	}
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if !config.Enabled {
		if _, err := container.Stdout().Write([]byte("Distributed cache is not enabled.\n")); err != nil {
			return err
		}
		if _, err := container.Stdout().Write([]byte("Configure it in ~/.config/buf/cache.yaml or set BUF_CACHE_DISTRIBUTED_ENABLED=true\n")); err != nil {
			return err
		}
		return nil
	}

	// Create backend
	backend, err := distributed.NewBackend(ctx, config.Backend)
	if err != nil {
		return fmt.Errorf("failed to create backend: %w", err)
	}

	// Get stats
	stats, err := backend.Stats(ctx)
	if err != nil {
		return fmt.Errorf("failed to get cache stats: %w", err)
	}

	// Format output
	var backendInfo string
	switch config.Backend.Type {
	case "s3":
		backendInfo = fmt.Sprintf("s3://%s/%s", config.Backend.Bucket, config.Backend.Prefix)
	case "gcs":
		backendInfo = fmt.Sprintf("gs://%s/%s", config.Backend.Bucket, config.Backend.Prefix)
	case "http":
		backendInfo = config.Backend.URL
	case "local":
		backendInfo = config.Backend.Path
	}

	output := fmt.Sprintf(`Backend: %s
Entries: %d
Size: %s
Hit rate: %.1f%%
`,
		backendInfo,
		stats.EntryCount,
		formatBytes(stats.Size),
		stats.HitRate(),
	)

	if _, err := container.Stdout().Write([]byte(output)); err != nil {
		return err
	}
	return nil
}

// formatBytes formats bytes into a human-readable string.
func formatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.1f GB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.1f MB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.1f KB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
