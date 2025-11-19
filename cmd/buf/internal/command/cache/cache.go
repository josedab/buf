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

// Package cache contains the cache command and its subcommands.
package cache

import (
	"buf.build/go/app/appcmd"
	"buf.build/go/app/appext"
	"github.com/bufbuild/buf/cmd/buf/internal/command/cache/cacheclean"
	"github.com/bufbuild/buf/cmd/buf/internal/command/cache/cachestatus"
	"github.com/bufbuild/buf/cmd/buf/internal/command/cache/cachewarm"
)

// NewCommand returns a new Command.
func NewCommand(
	name string,
	builder appext.SubCommandBuilder,
) *appcmd.Command {
	return &appcmd.Command{
		Use:   name,
		Short: "Manage the buf cache",
		Long:  "Manage the local plugin and module cache",
		SubCommands: []*appcmd.Command{
			cachestatus.NewCommand("status", builder),
			cacheclean.NewCommand("clean", builder),
			cachewarm.NewCommand("warm", builder),
		},
	}
}
