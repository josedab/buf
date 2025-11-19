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

package buflsp

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.lsp.dev/protocol"
)

func TestToLowerSnakeCase(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input    string
		expected string
	}{
		{"MyFieldName", "my_field_name"},
		{"myFieldName", "my_field_name"},
		{"fieldName", "field_name"},
		{"HTTPResponse", "h_t_t_p_response"},
		{"field", "field"},
		{"Field", "field"},
		{"ABC", "a_b_c"},
		{"already_snake_case", "already_snake_case"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			result := toLowerSnakeCase(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestToUpperSnakeCase(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input    string
		expected string
	}{
		{"MyEnumValue", "MY_ENUM_VALUE"},
		{"myEnumValue", "MY_ENUM_VALUE"},
		{"value", "VALUE"},
		{"VALUE", "v_a_l_u_e"},
		{"HTTPStatus", "H_T_T_P_STATUS"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			result := toUpperSnakeCase(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestToPascalCase(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input    string
		expected string
	}{
		{"my_message_name", "MyMessageName"},
		{"myMessageName", "Mymessagename"},
		{"message", "Message"},
		{"MESSAGE", "Message"},
		{"my_type", "MyType"},
		{"http_response", "HttpResponse"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			result := toPascalCase(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractIdentifierFromRange(t *testing.T) {
	t.Parallel()
	content := `syntax = "proto3";
message MyMessage {
  string myFieldName = 1;
}`

	tests := []struct {
		name     string
		content  string
		r        protocol.Range
		expected string
	}{
		{
			name:    "extract field name",
			content: content,
			r: protocol.Range{
				Start: protocol.Position{Line: 2, Character: 9},
				End:   protocol.Position{Line: 2, Character: 20},
			},
			expected: "myFieldName",
		},
		{
			name:    "extract message name",
			content: content,
			r: protocol.Range{
				Start: protocol.Position{Line: 1, Character: 8},
				End:   protocol.Position{Line: 1, Character: 17},
			},
			expected: "MyMessage",
		},
		{
			name:    "out of bounds line",
			content: content,
			r: protocol.Range{
				Start: protocol.Position{Line: 100, Character: 0},
				End:   protocol.Position{Line: 100, Character: 5},
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := extractIdentifierFromRange(tt.content, tt.r)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFieldLowerSnakeCaseAction(t *testing.T) {
	t.Parallel()

	action := &fieldLowerSnakeCaseAction{}
	content := `syntax = "proto3";
message MyMessage {
  string MyFieldName = 1;
}`

	diag := protocol.Diagnostic{
		Range: protocol.Range{
			Start: protocol.Position{Line: 2, Character: 9},
			End:   protocol.Position{Line: 2, Character: 20},
		},
		Code:    "FIELD_LOWER_SNAKE_CASE",
		Message: "Field name \"MyFieldName\" should be lower_snake_case",
	}

	doc := protocol.TextDocumentIdentifier{
		URI: "file:///test.proto",
	}

	result, err := action.Provide(context.Background(), doc, diag, content)
	require.NoError(t, err)

	assert.Equal(t, "Rename to 'my_field_name'", result.Title)
	assert.Equal(t, protocol.QuickFix, result.Kind)
	assert.True(t, result.IsPreferred)
	require.NotNil(t, result.Edit)
	require.Len(t, result.Edit.Changes, 1)
	edits := result.Edit.Changes[doc.URI]
	require.Len(t, edits, 1)
	assert.Equal(t, "my_field_name", edits[0].NewText)
}

func TestServiceSuffixAction(t *testing.T) {
	t.Parallel()

	action := &serviceSuffixAction{}
	content := `syntax = "proto3";
service User {
  rpc GetUser(GetUserRequest) returns (GetUserResponse);
}`

	diag := protocol.Diagnostic{
		Range: protocol.Range{
			Start: protocol.Position{Line: 1, Character: 8},
			End:   protocol.Position{Line: 1, Character: 12},
		},
		Code:    "SERVICE_SUFFIX",
		Message: "Service name \"User\" should end with \"Service\"",
	}

	doc := protocol.TextDocumentIdentifier{
		URI: "file:///test.proto",
	}

	result, err := action.Provide(context.Background(), doc, diag, content)
	require.NoError(t, err)

	assert.Equal(t, "Rename to 'UserService'", result.Title)
	assert.Equal(t, protocol.QuickFix, result.Kind)
	require.NotNil(t, result.Edit)
	edits := result.Edit.Changes[doc.URI]
	require.Len(t, edits, 1)
	assert.Equal(t, "UserService", edits[0].NewText)
}

func TestEnumValueUpperSnakeCaseAction(t *testing.T) {
	t.Parallel()

	action := &enumValueUpperSnakeCaseAction{}
	content := `syntax = "proto3";
enum Status {
  statusUnknown = 0;
}`

	diag := protocol.Diagnostic{
		Range: protocol.Range{
			Start: protocol.Position{Line: 2, Character: 2},
			End:   protocol.Position{Line: 2, Character: 15},
		},
		Code:    "ENUM_VALUE_UPPER_SNAKE_CASE",
		Message: "Enum value name \"statusUnknown\" should be UPPER_SNAKE_CASE",
	}

	doc := protocol.TextDocumentIdentifier{
		URI: "file:///test.proto",
	}

	result, err := action.Provide(context.Background(), doc, diag, content)
	require.NoError(t, err)

	assert.Equal(t, "Rename to 'STATUS_UNKNOWN'", result.Title)
	assert.Equal(t, protocol.QuickFix, result.Kind)
	require.NotNil(t, result.Edit)
	edits := result.Edit.Changes[doc.URI]
	require.Len(t, edits, 1)
	assert.Equal(t, "STATUS_UNKNOWN", edits[0].NewText)
}

func TestCodeActionNoChange(t *testing.T) {
	t.Parallel()

	action := &fieldLowerSnakeCaseAction{}
	content := `syntax = "proto3";
message MyMessage {
  string already_snake = 1;
}`

	diag := protocol.Diagnostic{
		Range: protocol.Range{
			Start: protocol.Position{Line: 2, Character: 9},
			End:   protocol.Position{Line: 2, Character: 22},
		},
		Code:    "FIELD_LOWER_SNAKE_CASE",
		Message: "Field name \"already_snake\" should be lower_snake_case",
	}

	doc := protocol.TextDocumentIdentifier{
		URI: "file:///test.proto",
	}

	result, err := action.Provide(context.Background(), doc, diag, content)
	require.NoError(t, err)

	// Should return empty action when no change is needed
	assert.Empty(t, result.Title)
}

func TestUTF16OffsetToByteOffset(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       string
		utf16Offset int
		expected    int
	}{
		{
			name:        "ASCII only",
			input:       "hello",
			utf16Offset: 3,
			expected:    3,
		},
		{
			name:        "with 2-byte UTF-8",
			input:       "héllo",
			utf16Offset: 3,
			expected:    4, // é is 2 bytes in UTF-8
		},
		{
			name:        "end of string",
			input:       "test",
			utf16Offset: 4,
			expected:    4,
		},
		{
			name:        "empty string",
			input:       "",
			utf16Offset: 0,
			expected:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := utf16OffsetToByteOffset(tt.input, tt.utf16Offset)
			assert.Equal(t, tt.expected, result)
		})
	}
}
