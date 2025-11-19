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
	"time"

	"buf.build/go/app/appcmd"
	"buf.build/go/app/appext"
	"github.com/bufbuild/buf/private/buf/bufcache/distributed"
	"github.com/spf13/pflag"
)

const (
	olderThanFlagName = "older-than"
	forceFlagName     = "force"
)

func newClearCommand(
	name string,
	builder appext.SubCommandBuilder,
) *appcmd.Command {
	flags := newClearFlags()
	return &appcmd.Command{
		Use:   name,
		Short: "Clear the distributed cache",
		Long:  "Remove entries from the distributed build cache.",
		Run: builder.NewRunFunc(
			func(ctx context.Context, container appext.Container) error {
				return runClear(ctx, container, flags)
			},
		),
		BindFlags: flags.Bind,
	}
}

type clearFlags struct {
	Config    string
	OlderThan string
	Force     bool
}

func newClearFlags() *clearFlags {
	return &clearFlags{}
}

func (f *clearFlags) Bind(flagSet *pflag.FlagSet) {
	flagSet.StringVar(
		&f.Config,
		configFlagName,
		"",
		"Path to cache configuration file",
	)
	flagSet.StringVar(
		&f.OlderThan,
		olderThanFlagName,
		"",
		"Only clear entries older than this duration (e.g., 30d, 7d, 24h)",
	)
	flagSet.BoolVar(
		&f.Force,
		forceFlagName,
		false,
		"Skip confirmation prompt",
	)
}

func runClear(
	ctx context.Context,
	container appext.Container,
	flags *clearFlags,
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
		return fmt.Errorf("distributed cache is not enabled")
	}

	// Parse older-than duration if provided
	var olderThan time.Duration
	if flags.OlderThan != "" {
		olderThan, err = parseDuration(flags.OlderThan)
		if err != nil {
			return fmt.Errorf("invalid duration: %w", err)
		}
	}

	// Create backend
	backend, err := distributed.NewBackend(ctx, config.Backend)
	if err != nil {
		return fmt.Errorf("failed to create backend: %w", err)
	}

	// Get current stats before clearing
	statsBefore, err := backend.Stats(ctx)
	if err != nil {
		return fmt.Errorf("failed to get cache stats: %w", err)
	}

	if statsBefore.EntryCount == 0 {
		if _, err := container.Stdout().Write([]byte("Cache is already empty.\n")); err != nil {
			return err
		}
		return nil
	}

	// Confirm unless forced
	if !flags.Force {
		var msg string
		if olderThan > 0 {
			msg = fmt.Sprintf("This will clear cache entries older than %s (%d entries, %s). Continue? [y/N] ",
				flags.OlderThan, statsBefore.EntryCount, formatBytes(statsBefore.Size))
		} else {
			msg = fmt.Sprintf("This will clear all cache entries (%d entries, %s). Continue? [y/N] ",
				statsBefore.EntryCount, formatBytes(statsBefore.Size))
		}
		if _, err := container.Stdout().Write([]byte(msg)); err != nil {
			return err
		}

		// Read confirmation
		var response string
		if _, err := fmt.Fscanln(container.Stdin(), &response); err != nil {
			return nil // User cancelled
		}
		if response != "y" && response != "Y" && response != "yes" && response != "Yes" {
			if _, err := container.Stdout().Write([]byte("Cancelled.\n")); err != nil {
				return err
			}
			return nil
		}
	}

	// For now, we can only clear for local backend with the Clear method
	// For S3/GCS/HTTP, we would need to implement listing and deleting
	if localBackend, ok := backend.(*distributed.LocalBackend); ok {
		if err := localBackend.Clear(ctx); err != nil {
			return fmt.Errorf("failed to clear cache: %w", err)
		}
	} else {
		// For other backends, we need to note this limitation
		if _, err := container.Stdout().Write([]byte("Note: Full cache clearing for remote backends requires manual intervention.\n")); err != nil {
			return err
		}
		if _, err := container.Stdout().Write([]byte("Please use your cloud provider's console or CLI to clear the bucket contents.\n")); err != nil {
			return err
		}
		return nil
	}

	if _, err := container.Stdout().Write([]byte("Cache cleared successfully.\n")); err != nil {
		return err
	}
	return nil
}

// parseDuration parses a duration string that can include days (e.g., "30d").
func parseDuration(s string) (time.Duration, error) {
	// Handle day format
	if len(s) > 0 && s[len(s)-1] == 'd' {
		var days int
		if _, err := fmt.Sscanf(s, "%dd", &days); err != nil {
			return 0, err
		}
		return time.Duration(days) * 24 * time.Hour, nil
	}

	return time.ParseDuration(s)
}
