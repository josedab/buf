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
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStructuredError_Error(t *testing.T) {
	t.Parallel()
	err := NewError(ModuleNotFound, "module \"buf.build/myorg/mymodule\" not found")
	assert.Equal(t, "BUFMD001: module \"buf.build/myorg/mymodule\" not found", err.Error())
}

func TestStructuredError_Unwrap(t *testing.T) {
	t.Parallel()
	underlying := errors.New("underlying error")
	err := NewError(
		ModuleNotFound,
		"module not found",
		WithUnderlying(underlying),
	)

	assert.Equal(t, underlying, err.Unwrap())
	assert.True(t, errors.Is(err, underlying))
}

func TestStructuredError_Format(t *testing.T) {
	t.Parallel()
	err := NewError(
		ModuleNotFound,
		"module \"buf.build/myorg/mymodule\" not found",
		WithDetail("Searched in registry: buf.build"),
		WithHelp("Run 'buf registry login' if this is a private module"),
	)

	formatted := err.Format()
	assert.Contains(t, formatted, "BUFMD001: module \"buf.build/myorg/mymodule\" not found")
	assert.Contains(t, formatted, "Searched in registry: buf.build")
	assert.Contains(t, formatted, "→ Run 'buf registry login' if this is a private module")
	assert.Contains(t, formatted, "→ See: https://buf.build/docs/errors/BUFMD001")
}

func TestStructuredError_Format_Multiline_Help(t *testing.T) {
	t.Parallel()
	err := NewError(
		ModuleNotFound,
		"module not found",
		WithHelp("Line 1\nLine 2\nLine 3"),
	)

	formatted := err.Format()
	assert.Contains(t, formatted, "→ Line 1")
	assert.Contains(t, formatted, "→ Line 2")
	assert.Contains(t, formatted, "→ Line 3")
}

func TestStructuredError_MarshalJSON(t *testing.T) {
	t.Parallel()
	err := NewError(
		ModuleNotFound,
		"module \"buf.build/myorg/mymodule\" not found",
		WithDetail("Searched in registry: buf.build"),
		WithHelp("Run 'buf registry login' if this is a private module"),
	)

	data, marshalErr := json.Marshal(err)
	require.NoError(t, marshalErr)

	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &result))

	assert.Equal(t, "BUFMD001", result["code"])
	assert.Equal(t, "module \"buf.build/myorg/mymodule\" not found", result["message"])
	assert.Equal(t, "Searched in registry: buf.build", result["detail"])
	assert.Equal(t, "Run 'buf registry login' if this is a private module", result["help"])
	assert.Equal(t, "https://buf.build/docs/errors/BUFMD001", result["url"])
}

func TestNewError(t *testing.T) {
	t.Parallel()
	err := NewError(AuthInvalidToken, "invalid token")
	assert.Equal(t, AuthInvalidToken, err.Code)
	assert.Equal(t, "invalid token", err.Message)
	assert.Empty(t, err.Detail)
	assert.Empty(t, err.Help)
}

func TestNewError_WithOptions(t *testing.T) {
	t.Parallel()
	err := NewError(
		AuthInvalidToken,
		"invalid token",
		WithDetail("Token format is incorrect"),
		WithHelp("Regenerate your token"),
	)
	assert.Equal(t, AuthInvalidToken, err.Code)
	assert.Equal(t, "invalid token", err.Message)
	assert.Equal(t, "Token format is incorrect", err.Detail)
	assert.Equal(t, "Regenerate your token", err.Help)
}

func TestNewErrorf(t *testing.T) {
	t.Parallel()
	err := NewErrorf(ModuleNotFound, "module %q not found", "buf.build/myorg/mymodule")
	assert.Equal(t, ModuleNotFound, err.Code)
	assert.Equal(t, "module \"buf.build/myorg/mymodule\" not found", err.Message)
}

func TestWithDetailf(t *testing.T) {
	t.Parallel()
	err := NewError(
		ModuleDependencyCycle,
		"circular dependency",
		WithDetailf("Cycle involves: %s -> %s", "moduleA", "moduleB"),
	)
	assert.Equal(t, "Cycle involves: moduleA -> moduleB", err.Detail)
}

func TestWithHelpf(t *testing.T) {
	t.Parallel()
	err := NewError(
		ConfigNotFound,
		"config not found",
		WithHelpf("Create %s in your project directory", "buf.yaml"),
	)
	assert.Equal(t, "Create buf.yaml in your project directory", err.Help)
}

func TestWrap(t *testing.T) {
	t.Parallel()
	underlying := errors.New("connection refused")
	err := Wrap(
		NetworkConnectionRefused,
		underlying,
		WithHelp("Check server status"),
	)

	assert.Equal(t, NetworkConnectionRefused, err.Code)
	assert.Equal(t, "connection refused", err.Message)
	assert.Equal(t, "Check server status", err.Help)
	assert.Equal(t, underlying, err.underlying)
}

func TestWrap_NilError(t *testing.T) {
	t.Parallel()
	err := Wrap(NetworkConnectionRefused, nil)
	assert.Nil(t, err)
}

func TestIsStructuredError(t *testing.T) {
	t.Parallel()

	structuredErr := NewError(AuthInvalidToken, "invalid token")
	assert.True(t, IsStructuredError(structuredErr))

	plainErr := errors.New("plain error")
	assert.False(t, IsStructuredError(plainErr))
}

func TestGetCode(t *testing.T) {
	t.Parallel()

	structuredErr := NewError(AuthInvalidToken, "invalid token")
	assert.Equal(t, AuthInvalidToken, GetCode(structuredErr))

	plainErr := errors.New("plain error")
	assert.Equal(t, Code(""), GetCode(plainErr))
}

func TestHasCode(t *testing.T) {
	t.Parallel()

	err := NewError(AuthInvalidToken, "invalid token")
	assert.True(t, HasCode(err, AuthInvalidToken))
	assert.False(t, HasCode(err, ModuleNotFound))

	plainErr := errors.New("plain error")
	assert.False(t, HasCode(plainErr, AuthInvalidToken))
}

func TestStructuredError_Format_NoOptionalFields(t *testing.T) {
	t.Parallel()
	err := NewError(ModuleNotFound, "module not found")
	formatted := err.Format()

	// Should still contain the error and URL
	assert.True(t, strings.HasPrefix(formatted, "BUFMD001: module not found\n"))
	assert.Contains(t, formatted, "→ See: https://buf.build/docs/errors/BUFMD001")

	// Should not have detail line
	lines := strings.Split(formatted, "\n")
	// First line: error message
	// Second line: URL
	// Empty line at the end
	// With no detail or help, we expect minimal output
	assert.GreaterOrEqual(t, len(lines), 2)
}
