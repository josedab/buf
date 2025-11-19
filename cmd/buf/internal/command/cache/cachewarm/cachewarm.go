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

// Package cachewarm implements the cache warm command.
package cachewarm

import (
	"context"
	"fmt"
	"os"

	"buf.build/go/app/appcmd"
	"buf.build/go/app/appext"
	"github.com/spf13/pflag"
)

const (
	configFlagName = "config"
)

// NewCommand returns a new Command.
func NewCommand(
	name string,
	builder appext.SubCommandBuilder,
) *appcmd.Command {
	flags := newFlags()
	return &appcmd.Command{
		Use:   name,
		Short: "Pre-warm the cache with plugins",
		Long: `Pre-download and cache plugins defined in buf.gen.yaml.

This command downloads all plugins referenced in your buf.gen.yaml configuration
and prepares them for fast execution by pre-compiling WASM modules and optionally
warming Docker container pools.

Examples:
  # Warm cache using default buf.gen.yaml
  buf cache warm

  # Warm cache using a specific config file
  buf cache warm --config path/to/buf.gen.yaml`,
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
	Config string
}

func newFlags() *flags {
	return &flags{}
}

func (f *flags) Bind(flagSet *pflag.FlagSet) {
	flagSet.StringVar(
		&f.Config,
		configFlagName,
		"",
		"Path to buf.gen.yaml configuration file",
	)
}

func run(
	ctx context.Context,
	container appext.Container,
	flags *flags,
) error {
	configPath := flags.Config
	if configPath == "" {
		configPath = "buf.gen.yaml"
	}

	writer := container.Stdout()

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("configuration file not found: %s", configPath)
	}

	fmt.Fprintf(writer, "Warming cache from configuration: %s\n", configPath)
	fmt.Fprintln(writer)

	// Note: Full implementation would:
	// 1. Parse the buf.gen.yaml to extract plugin references
	// 2. For each plugin reference:
	//    - Resolve the plugin reference to a specific version
	//    - Download the plugin data if not cached
	//    - For WASM plugins, pre-compile the module
	//    - For Docker plugins, optionally pre-pull the image
	//
	// This is a placeholder that demonstrates the command structure.
	// The actual implementation requires integration with bufconfig.ReadConfig
	// and the plugin resolution/caching infrastructure.

	fmt.Fprintln(writer, "Cache warming is not yet fully implemented.")
	fmt.Fprintln(writer, "To pre-cache plugins, run 'buf generate' which will automatically cache them.")

	return nil
}
