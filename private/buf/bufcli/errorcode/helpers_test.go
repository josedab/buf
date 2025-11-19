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
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewAuthInvalidTokenError(t *testing.T) {
	t.Parallel()
	err := NewAuthInvalidTokenError()
	assert.Equal(t, AuthInvalidToken, err.Code)
	assert.Contains(t, err.Message, "invalid")
	assert.NotEmpty(t, err.Help)
}

func TestNewAuthTokenExpiredError(t *testing.T) {
	t.Parallel()
	err := NewAuthTokenExpiredError()
	assert.Equal(t, AuthTokenExpired, err.Code)
	assert.Contains(t, err.Message, "expired")
	assert.NotEmpty(t, err.Help)
}

func TestNewAuthLoginRequiredError(t *testing.T) {
	t.Parallel()
	err := NewAuthLoginRequiredError("buf.build/myorg/mymodule")
	assert.Equal(t, AuthLoginRequired, err.Code)
	assert.Contains(t, err.Message, "buf.build/myorg/mymodule")
	assert.NotEmpty(t, err.Help)
}

func TestNewAuthPermissionDeniedError(t *testing.T) {
	t.Parallel()
	err := NewAuthPermissionDeniedError("delete module")
	assert.Equal(t, AuthPermissionDenied, err.Code)
	assert.Contains(t, err.Message, "delete module")
	assert.NotEmpty(t, err.Help)
}

func TestNewModuleNotFoundError(t *testing.T) {
	t.Parallel()
	err := NewModuleNotFoundError("buf.build/myorg/mymodule")
	assert.Equal(t, ModuleNotFound, err.Code)
	assert.Contains(t, err.Message, "buf.build/myorg/mymodule")
	assert.Contains(t, err.Message, "not found")
	assert.NotEmpty(t, err.Help)
	assert.Contains(t, err.Help, "buf registry login")
}

func TestNewModuleInvalidNameError(t *testing.T) {
	t.Parallel()
	err := NewModuleInvalidNameError("invalid-module-name")
	assert.Equal(t, ModuleInvalidName, err.Code)
	assert.Contains(t, err.Message, "invalid-module-name")
	assert.NotEmpty(t, err.Help)
}

func TestNewModuleDependencyCycleError(t *testing.T) {
	t.Parallel()
	modules := []string{"moduleA", "moduleB", "moduleC"}
	err := NewModuleDependencyCycleError(modules)
	assert.Equal(t, ModuleDependencyCycle, err.Code)
	assert.Contains(t, err.Message, "circular dependency")
	assert.Contains(t, err.Detail, "moduleA")
	assert.NotEmpty(t, err.Help)
}

func TestNewModuleVersionNotFoundError(t *testing.T) {
	t.Parallel()
	err := NewModuleVersionNotFoundError("buf.build/myorg/mymodule", "v1.0.0")
	assert.Equal(t, ModuleVersionNotFound, err.Code)
	assert.Contains(t, err.Message, "v1.0.0")
	assert.Contains(t, err.Message, "buf.build/myorg/mymodule")
	assert.NotEmpty(t, err.Help)
}

func TestNewModuleDependencyNotFoundError(t *testing.T) {
	t.Parallel()
	err := NewModuleDependencyNotFoundError("mymodule", "depmodule")
	assert.Equal(t, ModuleDependencyNotFound, err.Code)
	assert.Contains(t, err.Message, "mymodule")
	assert.Contains(t, err.Message, "depmodule")
	assert.NotEmpty(t, err.Help)
}

func TestNewWorkspaceNotFoundError(t *testing.T) {
	t.Parallel()
	err := NewWorkspaceNotFoundError()
	assert.Equal(t, WorkspaceNotFound, err.Code)
	assert.Contains(t, err.Message, "workspace")
	assert.Contains(t, err.Help, "buf.work.yaml")
}

func TestNewWorkspaceInvalidError(t *testing.T) {
	t.Parallel()
	err := NewWorkspaceInvalidError("duplicate module paths")
	assert.Equal(t, WorkspaceInvalid, err.Code)
	assert.Contains(t, err.Message, "duplicate module paths")
	assert.NotEmpty(t, err.Help)
}

func TestNewConfigNotFoundError(t *testing.T) {
	t.Parallel()
	err := NewConfigNotFoundError("buf.yaml")
	assert.Equal(t, ConfigNotFound, err.Code)
	assert.Contains(t, err.Message, "buf.yaml")
	assert.Contains(t, err.Help, "buf.yaml")
}

func TestNewConfigInvalidYAMLError(t *testing.T) {
	t.Parallel()
	underlying := errors.New("yaml: line 5: mapping values are not allowed here")
	err := NewConfigInvalidYAMLError("buf.yaml", underlying)
	assert.Equal(t, ConfigInvalidYAML, err.Code)
	assert.Contains(t, err.Message, "buf.yaml")
	assert.Equal(t, underlying, err.Unwrap())
	assert.NotEmpty(t, err.Help)
}

func TestNewConfigVersionInvalidError(t *testing.T) {
	t.Parallel()
	err := NewConfigVersionInvalidError("buf.yaml", "v3")
	assert.Equal(t, ConfigVersionInvalid, err.Code)
	assert.Contains(t, err.Message, "v3")
	assert.Contains(t, err.Message, "buf.yaml")
	assert.NotEmpty(t, err.Help)
}

func TestNewGeneratePluginFailedError(t *testing.T) {
	t.Parallel()
	underlying := errors.New("plugin crashed")
	err := NewGeneratePluginFailedError("protoc-gen-go", underlying)
	assert.Equal(t, GeneratePluginFailed, err.Code)
	assert.Contains(t, err.Message, "protoc-gen-go")
	assert.Equal(t, underlying, err.Unwrap())
	assert.NotEmpty(t, err.Help)
}

func TestNewGeneratePluginNotFoundError(t *testing.T) {
	t.Parallel()
	err := NewGeneratePluginNotFoundError("protoc-gen-go")
	assert.Equal(t, GeneratePluginNotFound, err.Code)
	assert.Contains(t, err.Message, "protoc-gen-go")
	assert.Contains(t, err.Help, "PATH")
}

func TestNewGenerateOutputError(t *testing.T) {
	t.Parallel()
	underlying := errors.New("permission denied")
	err := NewGenerateOutputError("/output/dir", underlying)
	assert.Equal(t, GenerateOutputError, err.Code)
	assert.Contains(t, err.Message, "/output/dir")
	assert.Equal(t, underlying, err.Unwrap())
	assert.NotEmpty(t, err.Help)
}

func TestNewNetworkDNSError(t *testing.T) {
	t.Parallel()
	underlying := errors.New("no such host")
	err := NewNetworkDNSError("buf.build", underlying)
	assert.Equal(t, NetworkDNSError, err.Code)
	assert.Contains(t, err.Message, "buf.build")
	assert.Equal(t, underlying, err.Unwrap())
	assert.NotEmpty(t, err.Help)
}

func TestNewNetworkTLSError(t *testing.T) {
	t.Parallel()
	underlying := errors.New("certificate verify failed")
	err := NewNetworkTLSError("buf.build", underlying)
	assert.Equal(t, NetworkTLSError, err.Code)
	assert.Contains(t, err.Message, "buf.build")
	assert.Equal(t, underlying, err.Unwrap())
	assert.NotEmpty(t, err.Help)
}

func TestNewNetworkTimeoutError(t *testing.T) {
	t.Parallel()
	err := NewNetworkTimeoutError("fetching module")
	assert.Equal(t, NetworkTimeout, err.Code)
	assert.Contains(t, err.Message, "fetching module")
	assert.NotEmpty(t, err.Help)
}

func TestNewNetworkConnectionRefusedError(t *testing.T) {
	t.Parallel()
	err := NewNetworkConnectionRefusedError("localhost:8080")
	assert.Equal(t, NetworkConnectionRefused, err.Code)
	assert.Contains(t, err.Message, "localhost:8080")
	assert.NotEmpty(t, err.Help)
}

func TestNewFilesystemPermissionDeniedError(t *testing.T) {
	t.Parallel()
	err := NewFilesystemPermissionDeniedError("/etc/passwd")
	assert.Equal(t, FilesystemPermissionDenied, err.Code)
	assert.Contains(t, err.Message, "/etc/passwd")
	assert.NotEmpty(t, err.Help)
}

func TestNewFilesystemNotFoundError(t *testing.T) {
	t.Parallel()
	err := NewFilesystemNotFoundError("/path/to/file")
	assert.Equal(t, FilesystemNotFound, err.Code)
	assert.Contains(t, err.Message, "/path/to/file")
	assert.NotEmpty(t, err.Help)
}

func TestNewFilesystemAlreadyExistsError(t *testing.T) {
	t.Parallel()
	err := NewFilesystemAlreadyExistsError("/path/to/file")
	assert.Equal(t, FilesystemAlreadyExists, err.Code)
	assert.Contains(t, err.Message, "/path/to/file")
	assert.Contains(t, err.Help, "--force")
}

func TestNewInternalError(t *testing.T) {
	t.Parallel()
	underlying := errors.New("nil pointer dereference")
	err := NewInternalError("parsing config", underlying)
	assert.Equal(t, InternalError, err.Code)
	assert.Contains(t, err.Message, "parsing config")
	assert.Equal(t, underlying, err.Unwrap())
	assert.Contains(t, err.Help, "bug")
}
