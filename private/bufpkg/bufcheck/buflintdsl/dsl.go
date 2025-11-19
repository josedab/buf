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

// Package buflintdsl provides a domain-specific language for defining custom lint rules.
package buflintdsl

import (
	"context"
	"fmt"
	"strings"

	"github.com/bufbuild/buf/private/bufpkg/bufanalysis"
	"github.com/bufbuild/buf/private/bufpkg/bufprotosource"
	"github.com/google/cel-go/cel"
)

// MatchKind represents the type of protobuf element to match.
type MatchKind string

const (
	// MatchKindMessage matches message descriptors.
	MatchKindMessage MatchKind = "message"
	// MatchKindField matches field descriptors.
	MatchKindField MatchKind = "field"
	// MatchKindEnum matches enum descriptors.
	MatchKindEnum MatchKind = "enum"
	// MatchKindEnumValue matches enum value descriptors.
	MatchKindEnumValue MatchKind = "enum_value"
	// MatchKindService matches service descriptors.
	MatchKindService MatchKind = "service"
	// MatchKindRPC matches RPC/method descriptors.
	MatchKindRPC MatchKind = "rpc"
	// MatchKindFile matches file descriptors.
	MatchKindFile MatchKind = "file"
	// MatchKindOneof matches oneof descriptors.
	MatchKindOneof MatchKind = "oneof"
)

// Severity represents the severity level of a rule.
type Severity string

const (
	// SeverityError represents an error severity.
	SeverityError Severity = "error"
	// SeverityWarning represents a warning severity.
	SeverityWarning Severity = "warning"
)

// RuleSpec defines a custom lint rule specification.
type RuleSpec struct {
	// ID is the unique identifier for the rule.
	ID string
	// Message is the template for the error message.
	// Supports template variables like {{field.name}}, {{message.name}}, etc.
	Message string
	// Severity is the severity level (error or warning).
	Severity Severity
	// Match defines what to match and the conditions.
	Match Match
}

// Match defines the matching criteria for a rule.
type Match struct {
	// Kind is the type of protobuf element to match.
	Kind MatchKind
	// Where contains CEL expressions that must all evaluate to true for a match.
	Where []string
}

// CompiledRule represents a compiled DSL rule ready for execution.
type CompiledRule struct {
	spec     RuleSpec
	programs []cel.Program
	env      *cel.Env
}

// CompileRule compiles a RuleSpec into a CompiledRule.
func CompileRule(spec RuleSpec) (*CompiledRule, error) {
	if spec.ID == "" {
		return nil, fmt.Errorf("rule ID is required")
	}
	if spec.Message == "" {
		return nil, fmt.Errorf("rule message is required for rule %s", spec.ID)
	}
	if spec.Match.Kind == "" {
		return nil, fmt.Errorf("match kind is required for rule %s", spec.ID)
	}

	// Validate match kind
	if !isValidMatchKind(spec.Match.Kind) {
		return nil, fmt.Errorf("invalid match kind %q for rule %s", spec.Match.Kind, spec.ID)
	}

	// Create CEL environment with declarations for the match kind
	env, err := newCELEnvForKind(spec.Match.Kind)
	if err != nil {
		return nil, fmt.Errorf("failed to create CEL environment for rule %s: %w", spec.ID, err)
	}

	// Compile all expressions
	var programs []cel.Program
	for i, expr := range spec.Match.Where {
		ast, issues := env.Compile(expr)
		if issues != nil && issues.Err() != nil {
			return nil, fmt.Errorf("failed to compile expression %d for rule %s: %w", i+1, spec.ID, issues.Err())
		}

		// Check that the expression returns a boolean
		if ast.OutputType() != cel.BoolType {
			return nil, fmt.Errorf("expression %d for rule %s must return bool, got %s", i+1, spec.ID, ast.OutputType())
		}

		prg, err := env.Program(ast)
		if err != nil {
			return nil, fmt.Errorf("failed to create program for expression %d in rule %s: %w", i+1, spec.ID, err)
		}
		programs = append(programs, prg)
	}

	return &CompiledRule{
		spec:     spec,
		programs: programs,
		env:      env,
	}, nil
}

// ID returns the rule ID.
func (r *CompiledRule) ID() string {
	return r.spec.ID
}

// Severity returns the rule severity.
func (r *CompiledRule) Severity() Severity {
	return r.spec.Severity
}

// Check checks a file and returns any annotations for violations.
func (r *CompiledRule) Check(ctx context.Context, file bufprotosource.File) ([]bufanalysis.FileAnnotation, error) {
	var annotations []bufanalysis.FileAnnotation

	items, err := itemsOfKind(file, r.spec.Match.Kind)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		matches, err := r.matches(item)
		if err != nil {
			return nil, err
		}
		if matches {
			annotation := r.createAnnotation(file, item)
			annotations = append(annotations, annotation)
		}
	}

	return annotations, nil
}

// matches evaluates all WHERE expressions against the item.
func (r *CompiledRule) matches(item any) (bool, error) {
	if len(r.programs) == 0 {
		return true, nil
	}

	activation := activationForItem(item, r.spec.Match.Kind)

	for _, prg := range r.programs {
		result, _, err := prg.Eval(activation)
		if err != nil {
			return false, fmt.Errorf("error evaluating expression: %w", err)
		}
		if result.Value() != true {
			return false, nil
		}
	}

	return true, nil
}

// createAnnotation creates a file annotation for a matched item.
func (r *CompiledRule) createAnnotation(file bufprotosource.File, item any) bufanalysis.FileAnnotation {
	message := r.formatMessage(item)

	var startLine, startColumn, endLine, endColumn int
	var fileInfo bufanalysis.FileInfo

	// Get location information from the item
	if locDesc, ok := item.(bufprotosource.LocationDescriptor); ok {
		if loc := locDesc.Location(); loc != nil {
			startLine = loc.StartLine()
			startColumn = loc.StartColumn()
			endLine = loc.EndLine()
			endColumn = loc.EndColumn()
		}
	}

	fileInfo = newFileInfo(file.Path(), file.ExternalPath())

	return bufanalysis.NewFileAnnotation(
		fileInfo,
		startLine,
		startColumn,
		endLine,
		endColumn,
		r.spec.ID,
		message,
		"", // pluginName
		"", // policyName
	)
}

// formatMessage formats the message template with values from the item.
func (r *CompiledRule) formatMessage(item any) string {
	message := r.spec.Message

	// Replace template variables based on item type
	switch v := item.(type) {
	case bufprotosource.Message:
		message = strings.ReplaceAll(message, "{{message.name}}", v.Name())
		message = strings.ReplaceAll(message, "{{message.full_name}}", v.FullName())
		message = strings.ReplaceAll(message, "{{len(fields)}}", fmt.Sprintf("%d", len(v.Fields())))
	case bufprotosource.Field:
		message = strings.ReplaceAll(message, "{{field.name}}", v.Name())
		message = strings.ReplaceAll(message, "{{field.full_name}}", v.FullName())
		if v.ParentMessage() != nil {
			message = strings.ReplaceAll(message, "{{message.name}}", v.ParentMessage().Name())
		}
	case bufprotosource.Enum:
		message = strings.ReplaceAll(message, "{{enum.name}}", v.Name())
		message = strings.ReplaceAll(message, "{{enum.full_name}}", v.FullName())
	case bufprotosource.EnumValue:
		message = strings.ReplaceAll(message, "{{value.name}}", v.Name())
		message = strings.ReplaceAll(message, "{{enum.name}}", v.Enum().Name())
	case bufprotosource.Service:
		message = strings.ReplaceAll(message, "{{service.name}}", v.Name())
		message = strings.ReplaceAll(message, "{{service.full_name}}", v.FullName())
	case bufprotosource.Method:
		message = strings.ReplaceAll(message, "{{rpc.name}}", v.Name())
		message = strings.ReplaceAll(message, "{{service.name}}", v.Service().Name())
	case bufprotosource.File:
		message = strings.ReplaceAll(message, "{{file.path}}", v.Path())
		message = strings.ReplaceAll(message, "{{file.package}}", v.Package())
	case bufprotosource.Oneof:
		message = strings.ReplaceAll(message, "{{oneof.name}}", v.Name())
		message = strings.ReplaceAll(message, "{{message.name}}", v.Message().Name())
	}

	return message
}

// isValidMatchKind checks if the match kind is valid.
func isValidMatchKind(kind MatchKind) bool {
	switch kind {
	case MatchKindMessage, MatchKindField, MatchKindEnum, MatchKindEnumValue,
		MatchKindService, MatchKindRPC, MatchKindFile, MatchKindOneof:
		return true
	default:
		return false
	}
}

// fileInfo implements bufanalysis.FileInfo.
type fileInfo struct {
	path         string
	externalPath string
}

func newFileInfo(path, externalPath string) *fileInfo {
	return &fileInfo{
		path:         path,
		externalPath: externalPath,
	}
}

func (f *fileInfo) Path() string {
	return f.path
}

func (f *fileInfo) ExternalPath() string {
	if f.externalPath != "" {
		return f.externalPath
	}
	return f.path
}
