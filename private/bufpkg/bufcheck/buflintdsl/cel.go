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

package buflintdsl

import (
	"strings"
	"unicode"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/google/cel-go/ext"
)

// newCELEnvForKind creates a CEL environment with declarations appropriate for the given match kind.
func newCELEnvForKind(kind MatchKind) (*cel.Env, error) {
	// Use CEL standard library for string functions
	// This provides: startsWith, endsWith, matches, contains, etc.
	standardLibs := []cel.EnvOption{
		ext.Strings(),
	}

	// Custom functions
	customFuncs := []cel.EnvOption{
		// Case conversion functions
		cel.Function("lower_snake_case",
			cel.Overload("lower_snake_case_string",
				[]*cel.Type{cel.StringType},
				cel.StringType,
				cel.UnaryBinding(func(str ref.Val) ref.Val {
					return types.String(toLowerSnakeCase(str.Value().(string)))
				}),
			),
		),
		cel.Function("upper_snake_case",
			cel.Overload("upper_snake_case_string",
				[]*cel.Type{cel.StringType},
				cel.StringType,
				cel.UnaryBinding(func(str ref.Val) ref.Val {
					return types.String(toUpperSnakeCase(str.Value().(string)))
				}),
			),
		),
		cel.Function("pascal_case",
			cel.Overload("pascal_case_string",
				[]*cel.Type{cel.StringType},
				cel.StringType,
				cel.UnaryBinding(func(str ref.Val) ref.Val {
					return types.String(toPascalCase(str.Value().(string)))
				}),
			),
		),
		cel.Function("camel_case",
			cel.Overload("camel_case_string",
				[]*cel.Type{cel.StringType},
				cel.StringType,
				cel.UnaryBinding(func(str ref.Val) ref.Val {
					return types.String(toCamelCase(str.Value().(string)))
				}),
			),
		),

		// Additional string utilities
		cel.Function("to_lower",
			cel.Overload("to_lower_string",
				[]*cel.Type{cel.StringType},
				cel.StringType,
				cel.UnaryBinding(func(str ref.Val) ref.Val {
					return types.String(strings.ToLower(str.Value().(string)))
				}),
			),
		),
		cel.Function("to_upper",
			cel.Overload("to_upper_string",
				[]*cel.Type{cel.StringType},
				cel.StringType,
				cel.UnaryBinding(func(str ref.Val) ref.Val {
					return types.String(strings.ToUpper(str.Value().(string)))
				}),
			),
		),
	}

	// Build variable declarations based on match kind
	var varDecls []cel.EnvOption
	switch kind {
	case MatchKindMessage:
		varDecls = messageDeclarations()
	case MatchKindField:
		varDecls = fieldDeclarations()
	case MatchKindEnum:
		varDecls = enumDeclarations()
	case MatchKindEnumValue:
		varDecls = enumValueDeclarations()
	case MatchKindService:
		varDecls = serviceDeclarations()
	case MatchKindRPC:
		varDecls = rpcDeclarations()
	case MatchKindFile:
		varDecls = fileDeclarations()
	case MatchKindOneof:
		varDecls = oneofDeclarations()
	}

	allOpts := append(standardLibs, customFuncs...)
	allOpts = append(allOpts, varDecls...)
	return cel.NewEnv(allOpts...)
}

// Variable declarations for each match kind

func messageDeclarations() []cel.EnvOption {
	return []cel.EnvOption{
		cel.Variable("name", cel.StringType),
		cel.Variable("full_name", cel.StringType),
		cel.Variable("nested_name", cel.StringType),
		cel.Variable("comment", cel.StringType),
		cel.Variable("deprecated", cel.BoolType),
		cel.Variable("is_map_entry", cel.BoolType),
		cel.Variable("field_count", cel.IntType),
		cel.Variable("oneof_count", cel.IntType),
		cel.Variable("nested_message_count", cel.IntType),
		cel.Variable("nested_enum_count", cel.IntType),
		cel.Variable("has_comment", cel.BoolType),
	}
}

func fieldDeclarations() []cel.EnvOption {
	return []cel.EnvOption{
		cel.Variable("name", cel.StringType),
		cel.Variable("full_name", cel.StringType),
		cel.Variable("number", cel.IntType),
		cel.Variable("field_type", cel.StringType),
		cel.Variable("type_name", cel.StringType),
		cel.Variable("label", cel.StringType),
		cel.Variable("comment", cel.StringType),
		cel.Variable("deprecated", cel.BoolType),
		cel.Variable("is_repeated", cel.BoolType),
		cel.Variable("is_map", cel.BoolType),
		cel.Variable("is_optional", cel.BoolType),
		cel.Variable("is_required", cel.BoolType),
		cel.Variable("has_comment", cel.BoolType),
		cel.Variable("json_name", cel.StringType),
		cel.Variable("message_name", cel.StringType),
		cel.Variable("has_default", cel.BoolType),
		cel.Variable("in_oneof", cel.BoolType),
	}
}

func enumDeclarations() []cel.EnvOption {
	return []cel.EnvOption{
		cel.Variable("name", cel.StringType),
		cel.Variable("full_name", cel.StringType),
		cel.Variable("nested_name", cel.StringType),
		cel.Variable("comment", cel.StringType),
		cel.Variable("deprecated", cel.BoolType),
		cel.Variable("value_count", cel.IntType),
		cel.Variable("has_comment", cel.BoolType),
		cel.Variable("allow_alias", cel.BoolType),
	}
}

func enumValueDeclarations() []cel.EnvOption {
	return []cel.EnvOption{
		cel.Variable("name", cel.StringType),
		cel.Variable("full_name", cel.StringType),
		cel.Variable("number", cel.IntType),
		cel.Variable("comment", cel.StringType),
		cel.Variable("deprecated", cel.BoolType),
		cel.Variable("has_comment", cel.BoolType),
		cel.Variable("enum_name", cel.StringType),
	}
}

func serviceDeclarations() []cel.EnvOption {
	return []cel.EnvOption{
		cel.Variable("name", cel.StringType),
		cel.Variable("full_name", cel.StringType),
		cel.Variable("comment", cel.StringType),
		cel.Variable("deprecated", cel.BoolType),
		cel.Variable("method_count", cel.IntType),
		cel.Variable("has_comment", cel.BoolType),
	}
}

func rpcDeclarations() []cel.EnvOption {
	return []cel.EnvOption{
		cel.Variable("name", cel.StringType),
		cel.Variable("full_name", cel.StringType),
		cel.Variable("comment", cel.StringType),
		cel.Variable("deprecated", cel.BoolType),
		cel.Variable("input_type", cel.StringType),
		cel.Variable("output_type", cel.StringType),
		cel.Variable("client_streaming", cel.BoolType),
		cel.Variable("server_streaming", cel.BoolType),
		cel.Variable("has_comment", cel.BoolType),
		cel.Variable("service_name", cel.StringType),
	}
}

func fileDeclarations() []cel.EnvOption {
	return []cel.EnvOption{
		cel.Variable("path", cel.StringType),
		cel.Variable("package", cel.StringType),
		cel.Variable("syntax", cel.StringType),
		cel.Variable("message_count", cel.IntType),
		cel.Variable("enum_count", cel.IntType),
		cel.Variable("service_count", cel.IntType),
		cel.Variable("deprecated", cel.BoolType),
	}
}

func oneofDeclarations() []cel.EnvOption {
	return []cel.EnvOption{
		cel.Variable("name", cel.StringType),
		cel.Variable("full_name", cel.StringType),
		cel.Variable("comment", cel.StringType),
		cel.Variable("has_comment", cel.BoolType),
		cel.Variable("field_count", cel.IntType),
		cel.Variable("message_name", cel.StringType),
	}
}

// Case conversion utilities

func toLowerSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				result.WriteRune('_')
			}
			result.WriteRune(unicode.ToLower(r))
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

func toUpperSnakeCase(s string) string {
	return strings.ToUpper(toLowerSnakeCase(s))
}

func toPascalCase(s string) string {
	var result strings.Builder
	capitalizeNext := true

	for _, r := range s {
		if r == '_' || r == '-' || r == ' ' {
			capitalizeNext = true
			continue
		}
		if capitalizeNext {
			result.WriteRune(unicode.ToUpper(r))
			capitalizeNext = false
		} else {
			result.WriteRune(unicode.ToLower(r))
		}
	}

	return result.String()
}

func toCamelCase(s string) string {
	pascal := toPascalCase(s)
	if len(pascal) == 0 {
		return pascal
	}
	return strings.ToLower(string(pascal[0])) + pascal[1:]
}
