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

// lintRuleToCode maps lint rule IDs to structured error codes.
var lintRuleToCode = map[string]Code{
	"COMMENT_ENUM":                        LintCommentEnum,
	"COMMENT_ENUM_VALUE":                  LintCommentEnumValue,
	"COMMENT_FIELD":                       LintCommentField,
	"COMMENT_MESSAGE":                     LintCommentMessage,
	"COMMENT_ONEOF":                       LintCommentOneof,
	"COMMENT_RPC":                         LintCommentRPC,
	"COMMENT_SERVICE":                     LintCommentService,
	"DIRECTORY_SAME_PACKAGE":              LintDirectorySamePackage,
	"ENUM_FIRST_VALUE_ZERO":               LintEnumFirstValueZero,
	"ENUM_NO_ALLOW_ALIAS":                 LintEnumNoAllowAlias,
	"ENUM_PASCAL_CASE":                    LintEnumPascalCase,
	"ENUM_VALUE_PREFIX":                   LintEnumValuePrefix,
	"ENUM_VALUE_UPPER_SNAKE_CASE":         LintEnumValueUpperSnakeCase,
	"ENUM_ZERO_VALUE_SUFFIX":              LintEnumZeroValueSuffix,
	"FIELD_LOWER_SNAKE_CASE":              LintFieldLowerSnakeCase,
	"FIELD_NOT_REQUIRED":                  LintFieldNotRequired,
	"FILE_LOWER_SNAKE_CASE":               LintFileLowerSnakeCase,
	"IMPORT_NO_PUBLIC":                    LintImportNoPublic,
	"IMPORT_NO_WEAK":                      LintImportNoWeak,
	"IMPORT_USED":                         LintImportUsed,
	"MESSAGE_PASCAL_CASE":                 LintMessagePascalCase,
	"ONEOF_LOWER_SNAKE_CASE":              LintOneofLowerSnakeCase,
	"PACKAGE_DEFINED":                     LintPackageDefined,
	"PACKAGE_DIRECTORY_MATCH":             LintPackageDirectoryMatch,
	"PACKAGE_LOWER_SNAKE_CASE":            LintPackageLowerSnakeCase,
	"PACKAGE_NO_IMPORT_CYCLE":             LintPackageNoImportCycle,
	"PACKAGE_SAME_DIRECTORY":              LintPackageSameDirectory,
	"PACKAGE_SAME_CSHARP_NAMESPACE":       LintPackageSameCsharpNamespace,
	"PACKAGE_SAME_GO_PACKAGE":             LintPackageSameGoPackage,
	"PACKAGE_SAME_JAVA_MULTIPLE_FILES":    LintPackageSameJavaMultipleFiles,
	"PACKAGE_SAME_JAVA_PACKAGE":           LintPackageSameJavaPackage,
	"PACKAGE_SAME_PHP_NAMESPACE":          LintPackageSamePhpNamespace,
	"PACKAGE_SAME_RUBY_PACKAGE":           LintPackageSameRubyPackage,
	"PACKAGE_SAME_SWIFT_PREFIX":           LintPackageSameSwiftPrefix,
	"PACKAGE_VERSION_SUFFIX":              LintPackageVersionSuffix,
	"PROTOVALIDATE":                       LintProtovalidate,
	"RPC_NO_CLIENT_STREAMING":             LintRPCNoClientStreaming,
	"RPC_NO_SERVER_STREAMING":             LintRPCNoServerStreaming,
	"RPC_PASCAL_CASE":                     LintRPCPascalCase,
	"RPC_REQUEST_RESPONSE_UNIQUE":         LintRPCRequestResponseUnique,
	"RPC_REQUEST_STANDARD_NAME":           LintRPCRequestStandardName,
	"RPC_RESPONSE_STANDARD_NAME":          LintRPCResponseStandardName,
	"SERVICE_PASCAL_CASE":                 LintServicePascalCase,
	"SERVICE_SUFFIX":                      LintServiceSuffix,
	"SYNTAX_SPECIFIED":                    LintSyntaxSpecified,
}

// breakingRuleToCode maps breaking rule IDs to structured error codes.
var breakingRuleToCode = map[string]Code{
	"ENUM_NO_DELETE":                                 BreakingEnumNoDelete,
	"ENUM_VALUE_NO_DELETE":                           BreakingEnumValueNoDelete,
	"ENUM_VALUE_NO_DELETE_UNLESS_NAME_RESERVED":      BreakingEnumValueNoDeleteUnlessNameReserved,
	"ENUM_VALUE_NO_DELETE_UNLESS_NUMBER_RESERVED":    BreakingEnumValueNoDeleteUnlessNumberReserved,
	"ENUM_VALUE_SAME_NAME":                           BreakingEnumValueSameName,
	"EXTENSION_MESSAGE_NO_DELETE":                    BreakingExtensionMessageNoDelete,
	"EXTENSION_NO_DELETE":                            BreakingExtensionNoDelete,
	"FIELD_NO_DELETE":                                BreakingFieldNoDelete,
	"FIELD_NO_DELETE_UNLESS_NAME_RESERVED":           BreakingFieldNoDeleteUnlessNameReserved,
	"FIELD_NO_DELETE_UNLESS_NUMBER_RESERVED":         BreakingFieldNoDeleteUnlessNumberReserved,
	"FIELD_SAME_CARDINALITY":                         BreakingFieldSameCardinality,
	"FIELD_SAME_CPP_STRING_TYPE":                     BreakingFieldSameCppStringType,
	"FIELD_SAME_CTYPE":                               BreakingFieldSameCtype,
	"FIELD_SAME_DEFAULT":                             BreakingFieldSameDefault,
	"FIELD_SAME_JSTYPE":                              BreakingFieldSameJstype,
	"FIELD_SAME_NAME":                                BreakingFieldSameName,
	"FIELD_SAME_ONEOF":                               BreakingFieldSameOneof,
	"FIELD_SAME_TYPE":                                BreakingFieldSameType,
	"FIELD_SAME_UTF8_VALIDATION":                     BreakingFieldSameUTF8Validation,
	"FIELD_WIRE_COMPATIBLE_CARDINALITY":              BreakingFieldWireCompatibleCardinality,
	"FIELD_WIRE_COMPATIBLE_TYPE":                     BreakingFieldWireCompatibleType,
	"FIELD_WIRE_JSON_COMPATIBLE_CARDINALITY":         BreakingFieldWireJSONCompatibleCardinality,
	"FIELD_WIRE_JSON_COMPATIBLE_TYPE":                BreakingFieldWireJSONCompatibleType,
	"FILE_NO_DELETE":                                 BreakingFileNoDelete,
	"FILE_SAME_CSHARP_NAMESPACE":                     BreakingFileSameCsharpNamespace,
	"FILE_SAME_GO_PACKAGE":                           BreakingFileSameGoPackage,
	"FILE_SAME_JAVA_MULTIPLE_FILES":                  BreakingFileSameJavaMultipleFiles,
	"FILE_SAME_JAVA_OUTER_CLASSNAME":                 BreakingFileSameJavaOuterClassname,
	"FILE_SAME_JAVA_PACKAGE":                         BreakingFileSameJavaPackage,
	"FILE_SAME_JAVA_STRING_CHECK_UTF8":               BreakingFileSameJavaStringCheckUtf8,
	"FILE_SAME_OBJC_CLASS_PREFIX":                    BreakingFileSameObjcClassPrefix,
	"FILE_SAME_OPTIMIZE_FOR":                         BreakingFileSameOptimizeFor,
	"FILE_SAME_PACKAGE":                              BreakingFileSamePackage,
	"FILE_SAME_PHP_CLASS_PREFIX":                     BreakingFileSamePhpClassPrefix,
	"FILE_SAME_PHP_METADATA_NAMESPACE":               BreakingFileSamePhpMetadataNamespace,
	"FILE_SAME_PHP_NAMESPACE":                        BreakingFileSamePhpNamespace,
	"FILE_SAME_RUBY_PACKAGE":                         BreakingFileSameRubyPackage,
	"FILE_SAME_SWIFT_PREFIX":                         BreakingFileSameSwiftPrefix,
	"FILE_SAME_SYNTAX":                               BreakingFileSameSyntax,
	"MESSAGE_NO_DELETE":                              BreakingMessageNoDelete,
	"MESSAGE_NO_REMOVE_STANDARD_DESCRIPTOR_ACCESSOR": BreakingMessageNoRemoveStandardDescriptorAccessor,
	"MESSAGE_SAME_JSON_FORMAT":                       BreakingMessageSameJSONFormat,
	"MESSAGE_SAME_MESSAGE_SET_WIRE_FORMAT":           BreakingMessageSameMessageSetWireFormat,
	"MESSAGE_SAME_REQUIRED_FIELDS":                   BreakingMessageSameRequiredFields,
	"ONEOF_NO_DELETE":                                BreakingOneofNoDelete,
	"PACKAGE_ENUM_NO_DELETE":                         BreakingPackageEnumNoDelete,
	"PACKAGE_EXTENSION_NO_DELETE":                    BreakingPackageExtensionNoDelete,
	"PACKAGE_MESSAGE_NO_DELETE":                      BreakingPackageMessageNoDelete,
	"PACKAGE_NO_DELETE":                              BreakingPackageNoDelete,
	"PACKAGE_SERVICE_NO_DELETE":                      BreakingPackageServiceNoDelete,
	"RESERVED_ENUM_NO_DELETE":                        BreakingReservedEnumNoDelete,
	"RESERVED_MESSAGE_NO_DELETE":                     BreakingReservedMessageNoDelete,
	"RPC_NO_DELETE":                                  BreakingRPCNoDelete,
	"RPC_SAME_CLIENT_STREAMING":                      BreakingRPCSameClientStreaming,
	"RPC_SAME_IDEMPOTENCY_LEVEL":                     BreakingRPCSameIdempotencyLevel,
	"RPC_SAME_REQUEST_TYPE":                          BreakingRPCSameRequestType,
	"RPC_SAME_RESPONSE_TYPE":                         BreakingRPCSameResponseType,
	"RPC_SAME_SERVER_STREAMING":                      BreakingRPCSameServerStreaming,
	"SERVICE_NO_DELETE":                              BreakingServiceNoDelete,
}

// LintRuleToCode returns the structured error code for a lint rule ID.
// Returns empty Code if the rule ID is not found.
func LintRuleToCode(ruleID string) Code {
	return lintRuleToCode[ruleID]
}

// BreakingRuleToCode returns the structured error code for a breaking rule ID.
// Returns empty Code if the rule ID is not found.
func BreakingRuleToCode(ruleID string) Code {
	return breakingRuleToCode[ruleID]
}

// RuleToCode returns the structured error code for any rule ID (lint or breaking).
// It first checks lint rules, then breaking rules.
// Returns empty Code if the rule ID is not found.
func RuleToCode(ruleID string) Code {
	if code := lintRuleToCode[ruleID]; code != "" {
		return code
	}
	return breakingRuleToCode[ruleID]
}

// IsLintRule returns true if the rule ID is a lint rule.
func IsLintRule(ruleID string) bool {
	_, ok := lintRuleToCode[ruleID]
	return ok
}

// IsBreakingRule returns true if the rule ID is a breaking rule.
func IsBreakingRule(ruleID string) bool {
	_, ok := breakingRuleToCode[ruleID]
	return ok
}

// AllLintRules returns a copy of the lint rule to code mapping.
func AllLintRules() map[string]Code {
	result := make(map[string]Code, len(lintRuleToCode))
	for k, v := range lintRuleToCode {
		result[k] = v
	}
	return result
}

// AllBreakingRules returns a copy of the breaking rule to code mapping.
func AllBreakingRules() map[string]Code {
	result := make(map[string]Code, len(breakingRuleToCode))
	for k, v := range breakingRuleToCode {
		result[k] = v
	}
	return result
}
