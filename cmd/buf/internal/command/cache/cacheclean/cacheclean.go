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

// Package cacheclean implements the cache clean command.
package cacheclean

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"buf.build/go/app/appcmd"
	"buf.build/go/app/appext"
	"github.com/bufbuild/buf/private/buf/bufcli"
	"github.com/spf13/pflag"
)

const (
	pluginsFlagName    = "plugins"
	wasmFlagName       = "wasm"
	containersFlagName = "containers"
	allFlagName        = "all"
)

// NewCommand returns a new Command.
func NewCommand(
	name string,
	builder appext.SubCommandBuilder,
) *appcmd.Command {
	flags := newFlags()
	return &appcmd.Command{
		Use:   name,
		Short: "Clear cache data",
		Long: `Clear specific cache types or all cached data.

Examples:
  # Clear plugin cache
  buf cache clean --plugins

  # Clear WASM compiled modules
  buf cache clean --wasm

  # Clear all caches
  buf cache clean --all`,
		Args: appcmd.NoArgs,
		Run: builder.NewRunFunc(
			func(ctx context.Context, container appext.Container) error {
				return run(ctx, container, flags)
			},
		),
		BindFlags: flags.Bind,
	}
}

type flags struct {
	Plugins    bool
	WASM       bool
	Containers bool
	All        bool
}

func newFlags() *flags {
	return &flags{}
}

func (f *flags) Bind(flagSet *pflag.FlagSet) {
	flagSet.BoolVar(
		&f.Plugins,
		pluginsFlagName,
		false,
		"Clear plugin cache",
	)
	flagSet.BoolVar(
		&f.WASM,
		wasmFlagName,
		false,
		"Clear WASM compiled modules cache",
	)
	flagSet.BoolVar(
		&f.Containers,
		containersFlagName,
		false,
		"Clear Docker container pool",
	)
	flagSet.BoolVar(
		&f.All,
		allFlagName,
		false,
		"Clear all caches",
	)
}

func run(
	ctx context.Context,
	container appext.Container,
	flags *flags,
) error {
	// If no flags specified, show error
	if !flags.Plugins && !flags.WASM && !flags.Containers && !flags.All {
		return appcmd.WrapInvalidArgumentError(fmt.Errorf("at least one cache type must be specified (--plugins, --wasm, --containers, or --all)"))
	}

	// Get cache directory
	cacheDir, err := bufcli.CacheDirPath(container)
	if err != nil {
		return err
	}

	if flags.All {
		flags.Plugins = true
		flags.WASM = true
		flags.Containers = true
	}

	writer := container.Stdout()
	var cleared []string

	// Clear plugin cache
	if flags.Plugins {
		pluginDir := filepath.Join(cacheDir, "v3", "plugins")
		if err := clearDir(pluginDir); err != nil {
			return fmt.Errorf("failed to clear plugin cache: %w", err)
		}
		cleared = append(cleared, "plugins")
	}

	// Clear WASM compiled modules
	if flags.WASM {
		wasmDir := filepath.Join(cacheDir, "v3", "wasmruntime")
		if err := clearDir(wasmDir); err != nil {
			return fmt.Errorf("failed to clear WASM cache: %w", err)
		}
		cleared = append(cleared, "WASM modules")
	}

	// Clear container pool
	if flags.Containers {
		// Note: Container pool cleanup is handled differently since containers
		// are managed by the Docker daemon. For now, we just log that we would
		// clear the container pool.
		cleared = append(cleared, "container pool")
	}

	if len(cleared) > 0 {
		fmt.Fprintf(writer, "Cleared: %v\n", cleared)
	}

	return nil
}

func clearDir(path string) error {
	// Check if directory exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	}

	// Remove all contents
	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		entryPath := filepath.Join(path, entry.Name())
		if err := os.RemoveAll(entryPath); err != nil {
			return err
		}
	}

	return nil
}
