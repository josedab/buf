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

import "fmt"

// Authentication error helpers

// NewAuthInvalidTokenError creates an error for invalid authentication tokens.
func NewAuthInvalidTokenError() *StructuredError {
	return NewError(
		AuthInvalidToken,
		"authentication token is invalid",
		WithHelp("Try logging in again with 'buf registry login'."),
	)
}

// NewAuthTokenExpiredError creates an error for expired authentication tokens.
func NewAuthTokenExpiredError() *StructuredError {
	return NewError(
		AuthTokenExpired,
		"authentication token has expired",
		WithHelp("Refresh your token with 'buf registry login'."),
	)
}

// NewAuthLoginRequiredError creates an error when authentication is required.
func NewAuthLoginRequiredError(resource string) *StructuredError {
	return NewError(
		AuthLoginRequired,
		fmt.Sprintf("authentication required to access %s", resource),
		WithHelp("Run 'buf registry login' to authenticate."),
	)
}

// NewAuthPermissionDeniedError creates an error for permission denied.
func NewAuthPermissionDeniedError(action string) *StructuredError {
	return NewError(
		AuthPermissionDenied,
		fmt.Sprintf("permission denied for %s", action),
		WithHelp("Check that you have the necessary permissions for this operation."),
	)
}

// Module error helpers

// NewModuleNotFoundError creates an error for module not found.
func NewModuleNotFoundError(moduleName string) *StructuredError {
	return NewError(
		ModuleNotFound,
		fmt.Sprintf("module %q not found", moduleName),
		WithHelp(fmt.Sprintf(
			"Ensure the module exists on the registry or check your dependencies.\nRun 'buf registry login' if this is a private module.",
		)),
	)
}

// NewModuleInvalidNameError creates an error for invalid module names.
func NewModuleInvalidNameError(moduleName string) *StructuredError {
	return NewError(
		ModuleInvalidName,
		fmt.Sprintf("module name %q is invalid", moduleName),
		WithHelp("Module names must be in the format 'registry/owner/repository'."),
	)
}

// NewModuleDependencyCycleError creates an error for circular dependencies.
func NewModuleDependencyCycleError(modules []string) *StructuredError {
	return NewError(
		ModuleDependencyCycle,
		"circular dependency detected in modules",
		WithDetailf("Cycle involves: %v", modules),
		WithHelp("Review your module dependencies and remove the circular reference."),
	)
}

// NewModuleVersionNotFoundError creates an error for version not found.
func NewModuleVersionNotFoundError(moduleName, version string) *StructuredError {
	return NewError(
		ModuleVersionNotFound,
		fmt.Sprintf("version %q not found for module %q", version, moduleName),
		WithHelp("Check that the version exists with 'buf registry commit list'."),
	)
}

// NewModuleDependencyNotFoundError creates an error for unresolved dependencies.
func NewModuleDependencyNotFoundError(moduleName, depName string) *StructuredError {
	return NewError(
		ModuleDependencyNotFound,
		fmt.Sprintf("dependency %q not found for module %q", depName, moduleName),
		WithHelp("Ensure the dependency is correctly specified in buf.yaml."),
	)
}

// Workspace error helpers

// NewWorkspaceNotFoundError creates an error for workspace not found.
func NewWorkspaceNotFoundError() *StructuredError {
	return NewError(
		WorkspaceNotFound,
		"no workspace found",
		WithHelp("Create a buf.work.yaml file in your project root to define a workspace."),
	)
}

// NewWorkspaceInvalidError creates an error for invalid workspace configuration.
func NewWorkspaceInvalidError(reason string) *StructuredError {
	return NewError(
		WorkspaceInvalid,
		fmt.Sprintf("workspace configuration is invalid: %s", reason),
		WithHelp("Check your buf.work.yaml for syntax errors."),
	)
}

// Config error helpers

// NewConfigNotFoundError creates an error for config file not found.
func NewConfigNotFoundError(configType string) *StructuredError {
	return NewError(
		ConfigNotFound,
		fmt.Sprintf("%s configuration file not found", configType),
		WithHelpf("Create a %s file in your project directory.", configType),
	)
}

// NewConfigInvalidYAMLError creates an error for invalid YAML syntax.
func NewConfigInvalidYAMLError(filename string, err error) *StructuredError {
	return NewError(
		ConfigInvalidYAML,
		fmt.Sprintf("invalid YAML syntax in %s", filename),
		WithUnderlying(err),
		WithHelp("Check your file for YAML syntax errors."),
	)
}

// NewConfigVersionInvalidError creates an error for invalid config version.
func NewConfigVersionInvalidError(filename, version string) *StructuredError {
	return NewError(
		ConfigVersionInvalid,
		fmt.Sprintf("invalid version %q in %s", version, filename),
		WithHelp("Use a supported configuration version."),
	)
}

// Generate error helpers

// NewGeneratePluginFailedError creates an error for plugin failures.
func NewGeneratePluginFailedError(pluginName string, err error) *StructuredError {
	return NewError(
		GeneratePluginFailed,
		fmt.Sprintf("code generation plugin %q failed", pluginName),
		WithUnderlying(err),
		WithHelp("Check the plugin output for details and ensure the plugin is correctly installed."),
	)
}

// NewGeneratePluginNotFoundError creates an error for missing plugins.
func NewGeneratePluginNotFoundError(pluginName string) *StructuredError {
	return NewError(
		GeneratePluginNotFound,
		fmt.Sprintf("code generation plugin %q not found", pluginName),
		WithHelp("Install the plugin or check that it is in your PATH."),
	)
}

// NewGenerateOutputError creates an error for output failures.
func NewGenerateOutputError(path string, err error) *StructuredError {
	return NewError(
		GenerateOutputError,
		fmt.Sprintf("failed to write generated output to %q", path),
		WithUnderlying(err),
		WithHelp("Check that you have write permissions to the output directory."),
	)
}

// Network error helpers

// NewNetworkDNSError creates an error for DNS resolution failures.
func NewNetworkDNSError(host string, err error) *StructuredError {
	return NewError(
		NetworkDNSError,
		fmt.Sprintf("DNS resolution failed for %q", host),
		WithUnderlying(err),
		WithHelp("Check your network connection and DNS configuration."),
	)
}

// NewNetworkTLSError creates an error for TLS/SSL failures.
func NewNetworkTLSError(host string, err error) *StructuredError {
	return NewError(
		NetworkTLSError,
		fmt.Sprintf("TLS connection failed to %q", host),
		WithUnderlying(err),
		WithHelp("Check if the server certificate is valid and your TLS configuration is correct."),
	)
}

// NewNetworkTimeoutError creates an error for network timeouts.
func NewNetworkTimeoutError(operation string) *StructuredError {
	return NewError(
		NetworkTimeout,
		fmt.Sprintf("network operation timed out: %s", operation),
		WithHelp("Check your network connection and try again. You may need to increase timeout settings."),
	)
}

// NewNetworkConnectionRefusedError creates an error for refused connections.
func NewNetworkConnectionRefusedError(host string) *StructuredError {
	return NewError(
		NetworkConnectionRefused,
		fmt.Sprintf("connection refused by %q", host),
		WithHelp("Check that the server is running and accessible."),
	)
}

// Filesystem error helpers

// NewFilesystemPermissionDeniedError creates an error for permission denied.
func NewFilesystemPermissionDeniedError(path string) *StructuredError {
	return NewError(
		FilesystemPermissionDenied,
		fmt.Sprintf("permission denied for %q", path),
		WithHelp("Check file permissions and try running with appropriate privileges."),
	)
}

// NewFilesystemNotFoundError creates an error for file not found.
func NewFilesystemNotFoundError(path string) *StructuredError {
	return NewError(
		FilesystemNotFound,
		fmt.Sprintf("file or directory not found: %q", path),
		WithHelp("Check that the path is correct and the file exists."),
	)
}

// NewFilesystemAlreadyExistsError creates an error for existing files.
func NewFilesystemAlreadyExistsError(path string) *StructuredError {
	return NewError(
		FilesystemAlreadyExists,
		fmt.Sprintf("file or directory already exists: %q", path),
		WithHelp("Use --force to overwrite or choose a different path."),
	)
}

// Internal error helpers

// NewInternalError creates an error for unexpected internal errors.
func NewInternalError(operation string, err error) *StructuredError {
	return NewError(
		InternalError,
		fmt.Sprintf("internal error during %s", operation),
		WithUnderlying(err),
		WithHelp("This is a bug. Please report it at https://github.com/bufbuild/buf/issues"),
	)
}
