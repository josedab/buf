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

// Package cachestatus implements the cache status command.
package cachestatus

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"buf.build/go/app/appcmd"
	"buf.build/go/app/appext"
	"github.com/bufbuild/buf/private/buf/bufcli"
	"github.com/bufbuild/buf/private/buf/bufprint"
	"github.com/spf13/pflag"
)

const (
	formatFlagName = "format"
)

// NewCommand returns a new Command.
func NewCommand(
	name string,
	builder appext.SubCommandBuilder,
) *appcmd.Command {
	flags := newFlags()
	return &appcmd.Command{
		Use:   name,
		Short: "Show cache status and statistics",
		Long:  "Display information about the plugin cache, WASM compiled modules, and container pool",
		Args:  appcmd.NoArgs,
		Run: builder.NewRunFunc(
			func(ctx context.Context, container appext.Container) error {
				return run(ctx, container, flags)
			},
		),
		BindFlags: flags.Bind,
	}
}

type flags struct {
	Format string
}

func newFlags() *flags {
	return &flags{}
}

func (f *flags) Bind(flagSet *pflag.FlagSet) {
	flagSet.StringVar(
		&f.Format,
		formatFlagName,
		bufprint.FormatText.String(),
		fmt.Sprintf(`The output format to use. Must be one of %s`, bufprint.AllFormatsString),
	)
}

func run(
	ctx context.Context,
	container appext.Container,
	flags *flags,
) error {
	format, err := bufprint.ParseFormat(flags.Format)
	if err != nil {
		return appcmd.WrapInvalidArgumentError(err)
	}

	// Get cache directory
	cacheDir, err := bufcli.CacheDirPath(container)
	if err != nil {
		return err
	}

	// Collect cache statistics
	status := &cacheStatus{}

	// Plugin cache
	pluginDir := filepath.Join(cacheDir, "v3", "plugins")
	pluginSize, pluginCount := getDirStats(pluginDir)
	status.PluginCacheSize = pluginSize
	status.PluginCount = pluginCount

	// WASM compiled modules
	wasmDir := filepath.Join(cacheDir, "v3", "wasmruntime")
	wasmSize, wasmCount := getDirStats(wasmDir)
	status.WASMCacheSize = wasmSize
	status.WASMModuleCount = wasmCount

	// Modules cache
	modulesDir := filepath.Join(cacheDir, "v3", "modules")
	modulesSize, modulesCount := getDirStats(modulesDir)
	status.ModulesCacheSize = modulesSize
	status.ModulesCount = modulesCount

	// Total cache size
	totalSize, _ := getDirStats(cacheDir)
	status.TotalCacheSize = totalSize
	status.CacheDirectory = cacheDir

	// Print status
	switch format {
	case bufprint.FormatText:
		return printTextStatus(container, status)
	case bufprint.FormatJSON:
		return printJSONStatus(container, status)
	default:
		return printTextStatus(container, status)
	}
}

type cacheStatus struct {
	CacheDirectory   string `json:"cache_directory"`
	TotalCacheSize   int64  `json:"total_cache_size"`
	PluginCacheSize  int64  `json:"plugin_cache_size"`
	PluginCount      int    `json:"plugin_count"`
	WASMCacheSize    int64  `json:"wasm_cache_size"`
	WASMModuleCount  int    `json:"wasm_module_count"`
	ModulesCacheSize int64  `json:"modules_cache_size"`
	ModulesCount     int    `json:"modules_count"`
}

func printTextStatus(container appext.Container, status *cacheStatus) error {
	writer := container.Stdout()
	fmt.Fprintf(writer, "Cache directory: %s\n", status.CacheDirectory)
	fmt.Fprintf(writer, "Total cache size: %s\n", formatBytes(status.TotalCacheSize))
	fmt.Fprintln(writer)
	fmt.Fprintf(writer, "Plugin cache: %s (%d plugins)\n", formatBytes(status.PluginCacheSize), status.PluginCount)
	fmt.Fprintf(writer, "WASM compiled: %s (%d modules)\n", formatBytes(status.WASMCacheSize), status.WASMModuleCount)
	fmt.Fprintf(writer, "Modules cache: %s (%d modules)\n", formatBytes(status.ModulesCacheSize), status.ModulesCount)
	return nil
}

func printJSONStatus(container appext.Container, status *cacheStatus) error {
	return bufprint.PrintJSON(container.Stdout(), status)
}

func getDirStats(path string) (size int64, count int) {
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}
		if !info.IsDir() {
			size += info.Size()
			count++
		}
		return nil
	})
	if err != nil {
		return 0, 0
	}
	return size, count
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
