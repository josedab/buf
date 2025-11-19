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
	"github.com/bufbuild/buf/private/buf/bufcli"
	"github.com/bufbuild/buf/private/buf/bufctl"
	"github.com/spf13/pflag"
)

func newWarmCommand(
	name string,
	builder appext.SubCommandBuilder,
) *appcmd.Command {
	flags := newWarmFlags()
	return &appcmd.Command{
		Use:   name + " <input>",
		Short: "Warm the distributed cache with the current project",
		Long:  "Build the current project and populate the distributed cache with the results.",
		Args:  appcmd.MaximumNArgs(1),
		Run: builder.NewRunFunc(
			func(ctx context.Context, container appext.Container) error {
				return runWarm(ctx, container, flags)
			},
		),
		BindFlags: flags.Bind,
	}
}

type warmFlags struct {
	Config          string
	DisableSymlinks bool
	InputHashtag    string
}

func newWarmFlags() *warmFlags {
	return &warmFlags{}
}

func (f *warmFlags) Bind(flagSet *pflag.FlagSet) {
	bufcli.BindInputHashtag(flagSet, &f.InputHashtag)
	bufcli.BindDisableSymlinks(flagSet, &f.DisableSymlinks, "disable-symlinks")
	flagSet.StringVar(
		&f.Config,
		configFlagName,
		"",
		"Path to cache configuration file",
	)
}

func runWarm(
	ctx context.Context,
	container appext.Container,
	flags *warmFlags,
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

	// Get input
	input, err := bufcli.GetInputValue(container, flags.InputHashtag, ".")
	if err != nil {
		return err
	}

	// Create controller and build image
	controller, err := bufcli.NewController(
		container,
		bufctl.WithDisableSymlinks(flags.DisableSymlinks),
	)
	if err != nil {
		return err
	}

	// Get workspace to collect files
	workspace, err := controller.GetWorkspace(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to get workspace: %w", err)
	}

	// Build image
	image, err := controller.GetImageForWorkspace(ctx, workspace)
	if err != nil {
		return fmt.Errorf("failed to build image: %w", err)
	}

	// Create backend and cache
	backend, err := distributed.NewBackend(ctx, config.Backend)
	if err != nil {
		return fmt.Errorf("failed to create backend: %w", err)
	}

	cache := distributed.NewCache(backend, bufcli.Version, container.Logger())

	// Collect files from workspace for cache key
	var files []distributed.File
	for _, module := range workspace.Modules() {
		moduleFileSet, err := module.ModuleFileSet()
		if err != nil {
			continue
		}
		for _, file := range moduleFileSet.Files() {
			content, err := file.Content()
			if err != nil {
				continue
			}
			files = append(files, distributed.File{
				Path:    file.Path(),
				Content: content,
			})
		}
	}

	// Store in cache
	options := distributed.CompileOptions{}
	if err := cache.PutImage(ctx, files, options, image); err != nil {
		return fmt.Errorf("failed to cache image: %w", err)
	}

	if _, err := container.Stdout().Write([]byte("Cache warmed successfully.\n")); err != nil {
		return err
	}
	return nil
}
