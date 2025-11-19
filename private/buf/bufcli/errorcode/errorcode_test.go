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

package errorcode

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCodeString(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		code     Code
		expected string
	}{
		{
			name:     "authentication error",
			code:     AuthInvalidToken,
			expected: "BUFAU001",
		},
		{
			name:     "module error",
			code:     ModuleNotFound,
			expected: "BUFMD001",
		},
		{
			name:     "lint rule",
			code:     LintFieldLowerSnakeCase,
			expected: "BUFLR015",
		},
		{
			name:     "breaking rule",
			code:     BreakingEnumNoDelete,
			expected: "BUFBR001",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, tt.code.String())
		})
	}
}

func TestCodeURL(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		code     Code
		expected string
	}{
		{
			name:     "authentication error",
			code:     AuthInvalidToken,
			expected: "https://buf.build/docs/errors/BUFAU001",
		},
		{
			name:     "module error",
			code:     ModuleNotFound,
			expected: "https://buf.build/docs/errors/BUFMD001",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, tt.code.URL())
		})
	}
}

func TestCodeCategory(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		code     Code
		expected string
	}{
		{
			name:     "authentication",
			code:     AuthInvalidToken,
			expected: "AU",
		},
		{
			name:     "module",
			code:     ModuleNotFound,
			expected: "MD",
		},
		{
			name:     "workspace",
			code:     WorkspaceNotFound,
			expected: "WS",
		},
		{
			name:     "config",
			code:     ConfigNotFound,
			expected: "CF",
		},
		{
			name:     "lint rule",
			code:     LintFieldLowerSnakeCase,
			expected: "LR",
		},
		{
			name:     "breaking rule",
			code:     BreakingEnumNoDelete,
			expected: "BR",
		},
		{
			name:     "generate",
			code:     GeneratePluginFailed,
			expected: "GN",
		},
		{
			name:     "network",
			code:     NetworkDNSError,
			expected: "NW",
		},
		{
			name:     "filesystem",
			code:     FilesystemNotFound,
			expected: "FS",
		},
		{
			name:     "internal",
			code:     InternalError,
			expected: "IN",
		},
		{
			name:     "empty code",
			code:     Code("BUF"),
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, tt.code.Category())
		})
	}
}

func TestCodeNumber(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		code     Code
		expected string
	}{
		{
			name:     "001",
			code:     AuthInvalidToken,
			expected: "001",
		},
		{
			name:     "002",
			code:     AuthTokenExpired,
			expected: "002",
		},
		{
			name:     "015",
			code:     LintFieldLowerSnakeCase,
			expected: "015",
		},
		{
			name:     "short code",
			code:     Code("BUF"),
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, tt.code.Number())
		})
	}
}

func TestCodeCategoryName(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		code     Code
		expected string
	}{
		{
			name:     "authentication",
			code:     AuthInvalidToken,
			expected: "Authentication",
		},
		{
			name:     "module",
			code:     ModuleNotFound,
			expected: "Module",
		},
		{
			name:     "workspace",
			code:     WorkspaceNotFound,
			expected: "Workspace",
		},
		{
			name:     "config",
			code:     ConfigNotFound,
			expected: "Config",
		},
		{
			name:     "lint rule",
			code:     LintFieldLowerSnakeCase,
			expected: "Lint Rule",
		},
		{
			name:     "breaking rule",
			code:     BreakingEnumNoDelete,
			expected: "Breaking Rule",
		},
		{
			name:     "generate",
			code:     GeneratePluginFailed,
			expected: "Generate",
		},
		{
			name:     "network",
			code:     NetworkDNSError,
			expected: "Network",
		},
		{
			name:     "filesystem",
			code:     FilesystemNotFound,
			expected: "Filesystem",
		},
		{
			name:     "internal",
			code:     InternalError,
			expected: "Internal",
		},
		{
			name:     "unknown category",
			code:     Code("BUFXX001"),
			expected: "Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, tt.code.CategoryName())
		})
	}
}
