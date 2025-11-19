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

// This file implements code actions (quick fixes) for LSP diagnostics.

package buflsp

import (
	"context"
	"regexp"
	"strings"
	"unicode"

	"go.lsp.dev/protocol"
)

// CodeActionProvider defines the interface for providing code actions for specific diagnostic codes.
type CodeActionProvider interface {
	// Provide generates a code action for the given diagnostic.
	Provide(
		ctx context.Context,
		doc protocol.TextDocumentIdentifier,
		diag protocol.Diagnostic,
		fileContent string,
	) (protocol.CodeAction, error)
}

// codeActionProviders maps diagnostic codes to their respective code action providers.
var codeActionProviders = map[string]CodeActionProvider{
	"FIELD_LOWER_SNAKE_CASE":      &fieldLowerSnakeCaseAction{},
	"ENUM_PASCAL_CASE":            &enumPascalCaseAction{},
	"ENUM_VALUE_UPPER_SNAKE_CASE": &enumValueUpperSnakeCaseAction{},
	"SERVICE_SUFFIX":              &serviceSuffixAction{},
	"MESSAGE_PASCAL_CASE":         &messagePascalCaseAction{},
	"RPC_PASCAL_CASE":             &rpcPascalCaseAction{},
	"ONEOF_LOWER_SNAKE_CASE":      &oneofLowerSnakeCaseAction{},
}

// TextDocumentCodeAction handles the textDocument/codeAction request.
func (s *server) TextDocumentCodeAction(
	ctx context.Context,
	params *protocol.CodeActionParams,
) ([]protocol.CodeAction, error) {
	file := s.fileManager.Get(params.TextDocument.URI)
	if file == nil {
		return nil, nil
	}

	var actions []protocol.CodeAction
	fileContent := file.file.Text()

	for _, diagnostic := range params.Context.Diagnostics {
		code, ok := diagnostic.Code.(string)
		if !ok {
			continue
		}

		provider, ok := codeActionProviders[code]
		if !ok {
			continue
		}

		action, err := provider.Provide(ctx, params.TextDocument, diagnostic, fileContent)
		if err != nil {
			s.lsp.logger.Debug("failed to provide code action", "error", err, "code", code)
			continue
		}

		// Only add non-empty actions
		if action.Title != "" {
			actions = append(actions, action)
		}
	}

	return actions, nil
}

// fieldLowerSnakeCaseAction provides quick fix for FIELD_LOWER_SNAKE_CASE violations.
type fieldLowerSnakeCaseAction struct{}

func (a *fieldLowerSnakeCaseAction) Provide(
	ctx context.Context,
	doc protocol.TextDocumentIdentifier,
	diag protocol.Diagnostic,
	fileContent string,
) (protocol.CodeAction, error) {
	oldName := extractIdentifierFromRange(fileContent, diag.Range)
	if oldName == "" {
		return protocol.CodeAction{}, nil
	}

	newName := toLowerSnakeCase(oldName)
	if newName == oldName {
		return protocol.CodeAction{}, nil
	}

	return protocol.CodeAction{
		Title: "Rename to '" + newName + "'",
		Kind:  protocol.QuickFix,
		Edit: &protocol.WorkspaceEdit{
			Changes: map[protocol.URI][]protocol.TextEdit{
				doc.URI: {
					{
						Range:   diag.Range,
						NewText: newName,
					},
				},
			},
		},
		IsPreferred: true,
		Diagnostics: []protocol.Diagnostic{diag},
	}, nil
}

// enumPascalCaseAction provides quick fix for ENUM_PASCAL_CASE violations.
type enumPascalCaseAction struct{}

func (a *enumPascalCaseAction) Provide(
	ctx context.Context,
	doc protocol.TextDocumentIdentifier,
	diag protocol.Diagnostic,
	fileContent string,
) (protocol.CodeAction, error) {
	oldName := extractIdentifierFromRange(fileContent, diag.Range)
	if oldName == "" {
		return protocol.CodeAction{}, nil
	}

	newName := toPascalCase(oldName)
	if newName == oldName {
		return protocol.CodeAction{}, nil
	}

	return protocol.CodeAction{
		Title: "Rename to '" + newName + "'",
		Kind:  protocol.QuickFix,
		Edit: &protocol.WorkspaceEdit{
			Changes: map[protocol.URI][]protocol.TextEdit{
				doc.URI: {
					{
						Range:   diag.Range,
						NewText: newName,
					},
				},
			},
		},
		IsPreferred: true,
		Diagnostics: []protocol.Diagnostic{diag},
	}, nil
}

// enumValueUpperSnakeCaseAction provides quick fix for ENUM_VALUE_UPPER_SNAKE_CASE violations.
type enumValueUpperSnakeCaseAction struct{}

func (a *enumValueUpperSnakeCaseAction) Provide(
	ctx context.Context,
	doc protocol.TextDocumentIdentifier,
	diag protocol.Diagnostic,
	fileContent string,
) (protocol.CodeAction, error) {
	oldName := extractIdentifierFromRange(fileContent, diag.Range)
	if oldName == "" {
		return protocol.CodeAction{}, nil
	}

	newName := toUpperSnakeCase(oldName)
	if newName == oldName {
		return protocol.CodeAction{}, nil
	}

	return protocol.CodeAction{
		Title: "Rename to '" + newName + "'",
		Kind:  protocol.QuickFix,
		Edit: &protocol.WorkspaceEdit{
			Changes: map[protocol.URI][]protocol.TextEdit{
				doc.URI: {
					{
						Range:   diag.Range,
						NewText: newName,
					},
				},
			},
		},
		IsPreferred: true,
		Diagnostics: []protocol.Diagnostic{diag},
	}, nil
}

// serviceSuffixAction provides quick fix for SERVICE_SUFFIX violations.
type serviceSuffixAction struct{}

func (a *serviceSuffixAction) Provide(
	ctx context.Context,
	doc protocol.TextDocumentIdentifier,
	diag protocol.Diagnostic,
	fileContent string,
) (protocol.CodeAction, error) {
	oldName := extractIdentifierFromRange(fileContent, diag.Range)
	if oldName == "" {
		return protocol.CodeAction{}, nil
	}

	// Add "Service" suffix if not present
	newName := oldName
	if !strings.HasSuffix(oldName, "Service") {
		newName = oldName + "Service"
	}
	if newName == oldName {
		return protocol.CodeAction{}, nil
	}

	return protocol.CodeAction{
		Title: "Rename to '" + newName + "'",
		Kind:  protocol.QuickFix,
		Edit: &protocol.WorkspaceEdit{
			Changes: map[protocol.URI][]protocol.TextEdit{
				doc.URI: {
					{
						Range:   diag.Range,
						NewText: newName,
					},
				},
			},
		},
		IsPreferred: true,
		Diagnostics: []protocol.Diagnostic{diag},
	}, nil
}

// messagePascalCaseAction provides quick fix for MESSAGE_PASCAL_CASE violations.
type messagePascalCaseAction struct{}

func (a *messagePascalCaseAction) Provide(
	ctx context.Context,
	doc protocol.TextDocumentIdentifier,
	diag protocol.Diagnostic,
	fileContent string,
) (protocol.CodeAction, error) {
	oldName := extractIdentifierFromRange(fileContent, diag.Range)
	if oldName == "" {
		return protocol.CodeAction{}, nil
	}

	newName := toPascalCase(oldName)
	if newName == oldName {
		return protocol.CodeAction{}, nil
	}

	return protocol.CodeAction{
		Title: "Rename to '" + newName + "'",
		Kind:  protocol.QuickFix,
		Edit: &protocol.WorkspaceEdit{
			Changes: map[protocol.URI][]protocol.TextEdit{
				doc.URI: {
					{
						Range:   diag.Range,
						NewText: newName,
					},
				},
			},
		},
		IsPreferred: true,
		Diagnostics: []protocol.Diagnostic{diag},
	}, nil
}

// rpcPascalCaseAction provides quick fix for RPC_PASCAL_CASE violations.
type rpcPascalCaseAction struct{}

func (a *rpcPascalCaseAction) Provide(
	ctx context.Context,
	doc protocol.TextDocumentIdentifier,
	diag protocol.Diagnostic,
	fileContent string,
) (protocol.CodeAction, error) {
	oldName := extractIdentifierFromRange(fileContent, diag.Range)
	if oldName == "" {
		return protocol.CodeAction{}, nil
	}

	newName := toPascalCase(oldName)
	if newName == oldName {
		return protocol.CodeAction{}, nil
	}

	return protocol.CodeAction{
		Title: "Rename to '" + newName + "'",
		Kind:  protocol.QuickFix,
		Edit: &protocol.WorkspaceEdit{
			Changes: map[protocol.URI][]protocol.TextEdit{
				doc.URI: {
					{
						Range:   diag.Range,
						NewText: newName,
					},
				},
			},
		},
		IsPreferred: true,
		Diagnostics: []protocol.Diagnostic{diag},
	}, nil
}

// oneofLowerSnakeCaseAction provides quick fix for ONEOF_LOWER_SNAKE_CASE violations.
type oneofLowerSnakeCaseAction struct{}

func (a *oneofLowerSnakeCaseAction) Provide(
	ctx context.Context,
	doc protocol.TextDocumentIdentifier,
	diag protocol.Diagnostic,
	fileContent string,
) (protocol.CodeAction, error) {
	oldName := extractIdentifierFromRange(fileContent, diag.Range)
	if oldName == "" {
		return protocol.CodeAction{}, nil
	}

	newName := toLowerSnakeCase(oldName)
	if newName == oldName {
		return protocol.CodeAction{}, nil
	}

	return protocol.CodeAction{
		Title: "Rename to '" + newName + "'",
		Kind:  protocol.QuickFix,
		Edit: &protocol.WorkspaceEdit{
			Changes: map[protocol.URI][]protocol.TextEdit{
				doc.URI: {
					{
						Range:   diag.Range,
						NewText: newName,
					},
				},
			},
		},
		IsPreferred: true,
		Diagnostics: []protocol.Diagnostic{diag},
	}, nil
}

// Helper functions for case conversion

// extractIdentifierFromRange extracts the identifier from the file content at the given range.
func extractIdentifierFromRange(content string, r protocol.Range) string {
	lines := strings.Split(content, "\n")
	if int(r.Start.Line) >= len(lines) {
		return ""
	}

	line := lines[r.Start.Line]
	if int(r.Start.Character) >= len(line) || int(r.End.Character) > len(line) {
		return ""
	}

	// Handle UTF-16 character positions
	start := utf16OffsetToByteOffset(line, int(r.Start.Character))
	end := utf16OffsetToByteOffset(line, int(r.End.Character))

	if start < 0 || end < 0 || start >= len(line) || end > len(line) || start >= end {
		return ""
	}

	return line[start:end]
}

// utf16OffsetToByteOffset converts a UTF-16 character offset to a byte offset.
func utf16OffsetToByteOffset(s string, utf16Offset int) int {
	byteOffset := 0
	utf16Count := 0

	for _, r := range s {
		if utf16Count >= utf16Offset {
			return byteOffset
		}

		runeLen := 1
		if r >= 0x10000 {
			runeLen = 2 // Surrogate pair in UTF-16
		}

		utf16Count += runeLen
		byteOffset += len(string(r))
	}

	if utf16Count >= utf16Offset {
		return byteOffset
	}

	return -1
}

// toLowerSnakeCase converts a string to lower_snake_case.
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

// toUpperSnakeCase converts a string to UPPER_SNAKE_CASE.
func toUpperSnakeCase(s string) string {
	// First convert to lower_snake_case, then uppercase
	snake := toLowerSnakeCase(s)
	return strings.ToUpper(snake)
}

// toPascalCase converts a string to PascalCase.
func toPascalCase(s string) string {
	// Split by underscores or find word boundaries
	var result strings.Builder
	capitalizeNext := true

	for _, r := range s {
		if r == '_' {
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

// Regex patterns for extracting names from diagnostic messages
var (
	fieldNameRegex   = regexp.MustCompile(`Field name "([^"]+)"`)
	enumNameRegex    = regexp.MustCompile(`Enum name "([^"]+)"`)
	serviceNameRegex = regexp.MustCompile(`Service name "([^"]+)"`)
	messageNameRegex = regexp.MustCompile(`Message name "([^"]+)"`)
	rpcNameRegex     = regexp.MustCompile(`RPC name "([^"]+)"`)
)
