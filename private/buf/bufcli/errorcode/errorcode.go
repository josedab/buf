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

// Package errorcode provides structured error codes for buf CLI errors.
//
// Error codes follow the format BUF{category}{number} where:
//   - BUF is the prefix for all buf errors
//   - category is a two-letter category code
//   - number is a three-digit error number
//
// Example: BUFAU001 for Authentication: Invalid token
package errorcode

import "fmt"

// Code represents a structured error code for buf CLI errors.
type Code string

// Authentication errors (AU)
const (
	// AuthInvalidToken indicates the authentication token is invalid.
	AuthInvalidToken Code = "BUFAU001"
	// AuthTokenExpired indicates the authentication token has expired.
	AuthTokenExpired Code = "BUFAU002"
	// AuthLoginRequired indicates authentication is required but not provided.
	AuthLoginRequired Code = "BUFAU003"
	// AuthPermissionDenied indicates the user lacks permission for the operation.
	AuthPermissionDenied Code = "BUFAU004"
	// AuthInvalidCredentials indicates the credentials are invalid.
	AuthInvalidCredentials Code = "BUFAU005"
)

// Module errors (MD)
const (
	// ModuleNotFound indicates the requested module could not be found.
	ModuleNotFound Code = "BUFMD001"
	// ModuleInvalidName indicates the module name is invalid.
	ModuleInvalidName Code = "BUFMD002"
	// ModuleDependencyCycle indicates a circular dependency was detected.
	ModuleDependencyCycle Code = "BUFMD003"
	// ModuleVersionNotFound indicates the requested module version was not found.
	ModuleVersionNotFound Code = "BUFMD004"
	// ModuleDigestMismatch indicates a module digest verification failure.
	ModuleDigestMismatch Code = "BUFMD005"
	// ModuleDependencyNotFound indicates a module dependency could not be resolved.
	ModuleDependencyNotFound Code = "BUFMD006"
)

// Workspace errors (WS)
const (
	// WorkspaceNotFound indicates no workspace was found.
	WorkspaceNotFound Code = "BUFWS001"
	// WorkspaceInvalid indicates the workspace configuration is invalid.
	WorkspaceInvalid Code = "BUFWS002"
	// WorkspaceModuleConflict indicates conflicting modules in the workspace.
	WorkspaceModuleConflict Code = "BUFWS003"
)

// Config errors (CF)
const (
	// ConfigNotFound indicates no configuration file was found.
	ConfigNotFound Code = "BUFCF001"
	// ConfigInvalidYAML indicates the configuration file has invalid YAML syntax.
	ConfigInvalidYAML Code = "BUFCF002"
	// ConfigVersionInvalid indicates the configuration file has an invalid version.
	ConfigVersionInvalid Code = "BUFCF003"
	// ConfigFieldInvalid indicates a configuration field has an invalid value.
	ConfigFieldInvalid Code = "BUFCF004"
	// ConfigDuplicate indicates duplicate configuration was found.
	ConfigDuplicate Code = "BUFCF005"
)

// Lint Rule errors (LR)
// These codes map to specific lint rules in buf.
const (
	// LintCommentEnum indicates missing comment on enum.
	LintCommentEnum Code = "BUFLR001"
	// LintCommentEnumValue indicates missing comment on enum value.
	LintCommentEnumValue Code = "BUFLR002"
	// LintCommentField indicates missing comment on field.
	LintCommentField Code = "BUFLR003"
	// LintCommentMessage indicates missing comment on message.
	LintCommentMessage Code = "BUFLR004"
	// LintCommentOneof indicates missing comment on oneof.
	LintCommentOneof Code = "BUFLR005"
	// LintCommentRPC indicates missing comment on RPC.
	LintCommentRPC Code = "BUFLR006"
	// LintCommentService indicates missing comment on service.
	LintCommentService Code = "BUFLR007"
	// LintDirectorySamePackage indicates files in directory have different packages.
	LintDirectorySamePackage Code = "BUFLR008"
	// LintEnumFirstValueZero indicates enum first value is not zero.
	LintEnumFirstValueZero Code = "BUFLR009"
	// LintEnumNoAllowAlias indicates enum uses allow_alias.
	LintEnumNoAllowAlias Code = "BUFLR010"
	// LintEnumPascalCase indicates enum name is not PascalCase.
	LintEnumPascalCase Code = "BUFLR011"
	// LintEnumValuePrefix indicates enum value lacks required prefix.
	LintEnumValuePrefix Code = "BUFLR012"
	// LintEnumValueUpperSnakeCase indicates enum value is not UPPER_SNAKE_CASE.
	LintEnumValueUpperSnakeCase Code = "BUFLR013"
	// LintEnumZeroValueSuffix indicates enum zero value lacks required suffix.
	LintEnumZeroValueSuffix Code = "BUFLR014"
	// LintFieldLowerSnakeCase indicates field name is not lower_snake_case.
	LintFieldLowerSnakeCase Code = "BUFLR015"
	// LintFieldNotRequired indicates field uses required keyword.
	LintFieldNotRequired Code = "BUFLR016"
	// LintFileLowerSnakeCase indicates filename is not lower_snake_case.
	LintFileLowerSnakeCase Code = "BUFLR017"
	// LintImportNoPublic indicates public import is used.
	LintImportNoPublic Code = "BUFLR018"
	// LintImportNoWeak indicates weak import is used.
	LintImportNoWeak Code = "BUFLR019"
	// LintImportUsed indicates unused import.
	LintImportUsed Code = "BUFLR020"
	// LintMessagePascalCase indicates message name is not PascalCase.
	LintMessagePascalCase Code = "BUFLR021"
	// LintOneofLowerSnakeCase indicates oneof name is not lower_snake_case.
	LintOneofLowerSnakeCase Code = "BUFLR022"
	// LintPackageDefined indicates package is not defined.
	LintPackageDefined Code = "BUFLR023"
	// LintPackageDirectoryMatch indicates package does not match directory.
	LintPackageDirectoryMatch Code = "BUFLR024"
	// LintPackageLowerSnakeCase indicates package is not lower_snake_case.
	LintPackageLowerSnakeCase Code = "BUFLR025"
	// LintPackageNoImportCycle indicates import cycle detected.
	LintPackageNoImportCycle Code = "BUFLR026"
	// LintPackageSameDirectory indicates package files in different directories.
	LintPackageSameDirectory Code = "BUFLR027"
	// LintPackageSameCsharpNamespace indicates inconsistent C# namespace.
	LintPackageSameCsharpNamespace Code = "BUFLR028"
	// LintPackageSameGoPackage indicates inconsistent Go package.
	LintPackageSameGoPackage Code = "BUFLR029"
	// LintPackageSameJavaMultipleFiles indicates inconsistent java_multiple_files.
	LintPackageSameJavaMultipleFiles Code = "BUFLR030"
	// LintPackageSameJavaPackage indicates inconsistent Java package.
	LintPackageSameJavaPackage Code = "BUFLR031"
	// LintPackageSamePhpNamespace indicates inconsistent PHP namespace.
	LintPackageSamePhpNamespace Code = "BUFLR032"
	// LintPackageSameRubyPackage indicates inconsistent Ruby package.
	LintPackageSameRubyPackage Code = "BUFLR033"
	// LintPackageSameSwiftPrefix indicates inconsistent Swift prefix.
	LintPackageSameSwiftPrefix Code = "BUFLR034"
	// LintPackageVersionSuffix indicates package lacks version suffix.
	LintPackageVersionSuffix Code = "BUFLR035"
	// LintProtovalidate indicates protovalidate rule violation.
	LintProtovalidate Code = "BUFLR036"
	// LintRPCNoClientStreaming indicates RPC uses client streaming.
	LintRPCNoClientStreaming Code = "BUFLR037"
	// LintRPCNoServerStreaming indicates RPC uses server streaming.
	LintRPCNoServerStreaming Code = "BUFLR038"
	// LintRPCPascalCase indicates RPC name is not PascalCase.
	LintRPCPascalCase Code = "BUFLR039"
	// LintRPCRequestResponseUnique indicates non-unique request/response types.
	LintRPCRequestResponseUnique Code = "BUFLR040"
	// LintRPCRequestStandardName indicates non-standard request type name.
	LintRPCRequestStandardName Code = "BUFLR041"
	// LintRPCResponseStandardName indicates non-standard response type name.
	LintRPCResponseStandardName Code = "BUFLR042"
	// LintServicePascalCase indicates service name is not PascalCase.
	LintServicePascalCase Code = "BUFLR043"
	// LintServiceSuffix indicates service lacks required suffix.
	LintServiceSuffix Code = "BUFLR044"
	// LintSyntaxSpecified indicates syntax is not specified.
	LintSyntaxSpecified Code = "BUFLR045"
)

// Breaking Rule errors (BR)
// These codes map to specific breaking change rules in buf.
const (
	// BreakingEnumNoDelete indicates enum was deleted.
	BreakingEnumNoDelete Code = "BUFBR001"
	// BreakingEnumValueNoDelete indicates enum value was deleted.
	BreakingEnumValueNoDelete Code = "BUFBR002"
	// BreakingEnumValueNoDeleteUnlessNameReserved indicates enum value deleted without reservation.
	BreakingEnumValueNoDeleteUnlessNameReserved Code = "BUFBR003"
	// BreakingEnumValueNoDeleteUnlessNumberReserved indicates enum value deleted without number reservation.
	BreakingEnumValueNoDeleteUnlessNumberReserved Code = "BUFBR004"
	// BreakingEnumValueSameName indicates enum value name changed.
	BreakingEnumValueSameName Code = "BUFBR005"
	// BreakingExtensionMessageNoDelete indicates extension message deleted.
	BreakingExtensionMessageNoDelete Code = "BUFBR006"
	// BreakingExtensionNoDelete indicates extension deleted.
	BreakingExtensionNoDelete Code = "BUFBR007"
	// BreakingFieldNoDelete indicates field was deleted.
	BreakingFieldNoDelete Code = "BUFBR008"
	// BreakingFieldNoDeleteUnlessNameReserved indicates field deleted without name reservation.
	BreakingFieldNoDeleteUnlessNameReserved Code = "BUFBR009"
	// BreakingFieldNoDeleteUnlessNumberReserved indicates field deleted without number reservation.
	BreakingFieldNoDeleteUnlessNumberReserved Code = "BUFBR010"
	// BreakingFieldSameCardinality indicates field cardinality changed.
	BreakingFieldSameCardinality Code = "BUFBR011"
	// BreakingFieldSameCppStringType indicates field C++ string type changed.
	BreakingFieldSameCppStringType Code = "BUFBR012"
	// BreakingFieldSameCtype indicates field ctype changed.
	BreakingFieldSameCtype Code = "BUFBR013"
	// BreakingFieldSameDefault indicates field default value changed.
	BreakingFieldSameDefault Code = "BUFBR014"
	// BreakingFieldSameJstype indicates field jstype changed.
	BreakingFieldSameJstype Code = "BUFBR015"
	// BreakingFieldSameName indicates field name changed.
	BreakingFieldSameName Code = "BUFBR016"
	// BreakingFieldSameOneof indicates field oneof changed.
	BreakingFieldSameOneof Code = "BUFBR017"
	// BreakingFieldSameType indicates field type changed.
	BreakingFieldSameType Code = "BUFBR018"
	// BreakingFieldSameUTF8Validation indicates field UTF-8 validation changed.
	BreakingFieldSameUTF8Validation Code = "BUFBR019"
	// BreakingFieldWireCompatibleCardinality indicates incompatible wire cardinality change.
	BreakingFieldWireCompatibleCardinality Code = "BUFBR020"
	// BreakingFieldWireCompatibleType indicates incompatible wire type change.
	BreakingFieldWireCompatibleType Code = "BUFBR021"
	// BreakingFieldWireJSONCompatibleCardinality indicates incompatible JSON cardinality.
	BreakingFieldWireJSONCompatibleCardinality Code = "BUFBR022"
	// BreakingFieldWireJSONCompatibleType indicates incompatible JSON type.
	BreakingFieldWireJSONCompatibleType Code = "BUFBR023"
	// BreakingFileNoDelete indicates file was deleted.
	BreakingFileNoDelete Code = "BUFBR024"
	// BreakingFileSameCsharpNamespace indicates C# namespace changed.
	BreakingFileSameCsharpNamespace Code = "BUFBR025"
	// BreakingFileSameGoPackage indicates Go package changed.
	BreakingFileSameGoPackage Code = "BUFBR026"
	// BreakingFileSameJavaMultipleFiles indicates java_multiple_files changed.
	BreakingFileSameJavaMultipleFiles Code = "BUFBR027"
	// BreakingFileSameJavaOuterClassname indicates Java outer classname changed.
	BreakingFileSameJavaOuterClassname Code = "BUFBR028"
	// BreakingFileSameJavaPackage indicates Java package changed.
	BreakingFileSameJavaPackage Code = "BUFBR029"
	// BreakingFileSameJavaStringCheckUtf8 indicates java_string_check_utf8 changed.
	BreakingFileSameJavaStringCheckUtf8 Code = "BUFBR030"
	// BreakingFileSameObjcClassPrefix indicates Objective-C class prefix changed.
	BreakingFileSameObjcClassPrefix Code = "BUFBR031"
	// BreakingFileSameOptimizeFor indicates optimize_for changed.
	BreakingFileSameOptimizeFor Code = "BUFBR032"
	// BreakingFileSamePackage indicates package changed.
	BreakingFileSamePackage Code = "BUFBR033"
	// BreakingFileSamePhpClassPrefix indicates PHP class prefix changed.
	BreakingFileSamePhpClassPrefix Code = "BUFBR034"
	// BreakingFileSamePhpMetadataNamespace indicates PHP metadata namespace changed.
	BreakingFileSamePhpMetadataNamespace Code = "BUFBR035"
	// BreakingFileSamePhpNamespace indicates PHP namespace changed.
	BreakingFileSamePhpNamespace Code = "BUFBR036"
	// BreakingFileSameRubyPackage indicates Ruby package changed.
	BreakingFileSameRubyPackage Code = "BUFBR037"
	// BreakingFileSameSwiftPrefix indicates Swift prefix changed.
	BreakingFileSameSwiftPrefix Code = "BUFBR038"
	// BreakingFileSameSyntax indicates syntax changed.
	BreakingFileSameSyntax Code = "BUFBR039"
	// BreakingMessageNoDelete indicates message was deleted.
	BreakingMessageNoDelete Code = "BUFBR040"
	// BreakingMessageNoRemoveStandardDescriptorAccessor indicates accessor removed.
	BreakingMessageNoRemoveStandardDescriptorAccessor Code = "BUFBR041"
	// BreakingMessageSameJSONFormat indicates JSON format changed.
	BreakingMessageSameJSONFormat Code = "BUFBR042"
	// BreakingMessageSameMessageSetWireFormat indicates message_set_wire_format changed.
	BreakingMessageSameMessageSetWireFormat Code = "BUFBR043"
	// BreakingMessageSameRequiredFields indicates required fields changed.
	BreakingMessageSameRequiredFields Code = "BUFBR044"
	// BreakingOneofNoDelete indicates oneof was deleted.
	BreakingOneofNoDelete Code = "BUFBR045"
	// BreakingPackageEnumNoDelete indicates enum deleted from package.
	BreakingPackageEnumNoDelete Code = "BUFBR046"
	// BreakingPackageExtensionNoDelete indicates extension deleted from package.
	BreakingPackageExtensionNoDelete Code = "BUFBR047"
	// BreakingPackageMessageNoDelete indicates message deleted from package.
	BreakingPackageMessageNoDelete Code = "BUFBR048"
	// BreakingPackageNoDelete indicates package was deleted.
	BreakingPackageNoDelete Code = "BUFBR049"
	// BreakingPackageServiceNoDelete indicates service deleted from package.
	BreakingPackageServiceNoDelete Code = "BUFBR050"
	// BreakingReservedEnumNoDelete indicates reserved enum deleted.
	BreakingReservedEnumNoDelete Code = "BUFBR051"
	// BreakingReservedMessageNoDelete indicates reserved message deleted.
	BreakingReservedMessageNoDelete Code = "BUFBR052"
	// BreakingRPCNoDelete indicates RPC was deleted.
	BreakingRPCNoDelete Code = "BUFBR053"
	// BreakingRPCSameClientStreaming indicates client streaming changed.
	BreakingRPCSameClientStreaming Code = "BUFBR054"
	// BreakingRPCSameIdempotencyLevel indicates idempotency level changed.
	BreakingRPCSameIdempotencyLevel Code = "BUFBR055"
	// BreakingRPCSameRequestType indicates request type changed.
	BreakingRPCSameRequestType Code = "BUFBR056"
	// BreakingRPCSameResponseType indicates response type changed.
	BreakingRPCSameResponseType Code = "BUFBR057"
	// BreakingRPCSameServerStreaming indicates server streaming changed.
	BreakingRPCSameServerStreaming Code = "BUFBR058"
	// BreakingServiceNoDelete indicates service was deleted.
	BreakingServiceNoDelete Code = "BUFBR059"
)

// Generate errors (GN)
const (
	// GeneratePluginFailed indicates a code generation plugin failed.
	GeneratePluginFailed Code = "BUFGN001"
	// GenerateTemplateInvalid indicates the generation template is invalid.
	GenerateTemplateInvalid Code = "BUFGN002"
	// GenerateOutputError indicates an error writing generated output.
	GenerateOutputError Code = "BUFGN003"
	// GeneratePluginNotFound indicates the generation plugin was not found.
	GeneratePluginNotFound Code = "BUFGN004"
)

// Network errors (NW)
const (
	// NetworkDNSError indicates a DNS resolution failure.
	NetworkDNSError Code = "BUFNW001"
	// NetworkTLSError indicates a TLS/SSL connection failure.
	NetworkTLSError Code = "BUFNW002"
	// NetworkTimeout indicates a network operation timed out.
	NetworkTimeout Code = "BUFNW003"
	// NetworkConnectionRefused indicates the connection was refused.
	NetworkConnectionRefused Code = "BUFNW004"
	// NetworkUnreachable indicates the network is unreachable.
	NetworkUnreachable Code = "BUFNW005"
)

// Filesystem errors (FS)
const (
	// FilesystemPermissionDenied indicates permission was denied.
	FilesystemPermissionDenied Code = "BUFFS001"
	// FilesystemNotFound indicates the file or directory was not found.
	FilesystemNotFound Code = "BUFFS002"
	// FilesystemAlreadyExists indicates the file or directory already exists.
	FilesystemAlreadyExists Code = "BUFFS003"
	// FilesystemDiskFull indicates the disk is full.
	FilesystemDiskFull Code = "BUFFS004"
	// FilesystemInvalidPath indicates an invalid file path.
	FilesystemInvalidPath Code = "BUFFS005"
)

// Internal errors (IN)
const (
	// InternalError indicates an unexpected internal error.
	InternalError Code = "BUFIN001"
	// InternalPanic indicates an internal panic occurred.
	InternalPanic Code = "BUFIN002"
	// InternalDataCorruption indicates internal data corruption.
	InternalDataCorruption Code = "BUFIN003"
)

// String returns the string representation of the error code.
func (c Code) String() string {
	return string(c)
}

// URL returns the documentation URL for the error code.
func (c Code) URL() string {
	return fmt.Sprintf("https://buf.build/docs/errors/%s", c)
}

// Category returns the category portion of the error code (e.g., "AU" for BUFAU001).
func (c Code) Category() string {
	if len(c) < 5 {
		return ""
	}
	return string(c[3:5])
}

// Number returns the numeric portion of the error code (e.g., "001" for BUFAU001).
func (c Code) Number() string {
	if len(c) < 8 {
		return ""
	}
	return string(c[5:8])
}

// CategoryName returns the human-readable name of the error category.
func (c Code) CategoryName() string {
	switch c.Category() {
	case "AU":
		return "Authentication"
	case "MD":
		return "Module"
	case "WS":
		return "Workspace"
	case "CF":
		return "Config"
	case "LR":
		return "Lint Rule"
	case "BR":
		return "Breaking Rule"
	case "GN":
		return "Generate"
	case "NW":
		return "Network"
	case "FS":
		return "Filesystem"
	case "IN":
		return "Internal"
	default:
		return "Unknown"
	}
}
