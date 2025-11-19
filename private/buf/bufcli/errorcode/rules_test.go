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

func TestLintRuleToCode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		ruleID   string
		expected Code
	}{
		{
			name:     "field lower snake case",
			ruleID:   "FIELD_LOWER_SNAKE_CASE",
			expected: LintFieldLowerSnakeCase,
		},
		{
			name:     "enum pascal case",
			ruleID:   "ENUM_PASCAL_CASE",
			expected: LintEnumPascalCase,
		},
		{
			name:     "message pascal case",
			ruleID:   "MESSAGE_PASCAL_CASE",
			expected: LintMessagePascalCase,
		},
		{
			name:     "service suffix",
			ruleID:   "SERVICE_SUFFIX",
			expected: LintServiceSuffix,
		},
		{
			name:     "unknown rule",
			ruleID:   "UNKNOWN_RULE",
			expected: "",
		},
		{
			name:     "empty rule",
			ruleID:   "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, LintRuleToCode(tt.ruleID))
		})
	}
}

func TestBreakingRuleToCode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		ruleID   string
		expected Code
	}{
		{
			name:     "enum no delete",
			ruleID:   "ENUM_NO_DELETE",
			expected: BreakingEnumNoDelete,
		},
		{
			name:     "field no delete",
			ruleID:   "FIELD_NO_DELETE",
			expected: BreakingFieldNoDelete,
		},
		{
			name:     "field same type",
			ruleID:   "FIELD_SAME_TYPE",
			expected: BreakingFieldSameType,
		},
		{
			name:     "rpc no delete",
			ruleID:   "RPC_NO_DELETE",
			expected: BreakingRPCNoDelete,
		},
		{
			name:     "unknown rule",
			ruleID:   "UNKNOWN_RULE",
			expected: "",
		},
		{
			name:     "empty rule",
			ruleID:   "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, BreakingRuleToCode(tt.ruleID))
		})
	}
}

func TestRuleToCode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		ruleID   string
		expected Code
	}{
		{
			name:     "lint rule",
			ruleID:   "FIELD_LOWER_SNAKE_CASE",
			expected: LintFieldLowerSnakeCase,
		},
		{
			name:     "breaking rule",
			ruleID:   "ENUM_NO_DELETE",
			expected: BreakingEnumNoDelete,
		},
		{
			name:     "unknown rule",
			ruleID:   "UNKNOWN_RULE",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, RuleToCode(tt.ruleID))
		})
	}
}

func TestIsLintRule(t *testing.T) {
	t.Parallel()
	assert.True(t, IsLintRule("FIELD_LOWER_SNAKE_CASE"))
	assert.True(t, IsLintRule("ENUM_PASCAL_CASE"))
	assert.False(t, IsLintRule("ENUM_NO_DELETE"))
	assert.False(t, IsLintRule("UNKNOWN_RULE"))
}

func TestIsBreakingRule(t *testing.T) {
	t.Parallel()
	assert.True(t, IsBreakingRule("ENUM_NO_DELETE"))
	assert.True(t, IsBreakingRule("FIELD_SAME_TYPE"))
	assert.False(t, IsBreakingRule("FIELD_LOWER_SNAKE_CASE"))
	assert.False(t, IsBreakingRule("UNKNOWN_RULE"))
}

func TestAllLintRules(t *testing.T) {
	t.Parallel()
	rules := AllLintRules()

	// Verify it's a copy
	rules["TEST_RULE"] = "TEST"
	assert.Equal(t, Code(""), LintRuleToCode("TEST_RULE"))

	// Verify some known rules are present
	assert.Equal(t, LintFieldLowerSnakeCase, rules["FIELD_LOWER_SNAKE_CASE"])
	assert.Equal(t, LintEnumPascalCase, rules["ENUM_PASCAL_CASE"])
	assert.Equal(t, LintServiceSuffix, rules["SERVICE_SUFFIX"])

	// Verify count is reasonable (we defined 45 lint rules)
	assert.GreaterOrEqual(t, len(rules), 40)
}

func TestAllBreakingRules(t *testing.T) {
	t.Parallel()
	rules := AllBreakingRules()

	// Verify it's a copy
	rules["TEST_RULE"] = "TEST"
	assert.Equal(t, Code(""), BreakingRuleToCode("TEST_RULE"))

	// Verify some known rules are present
	assert.Equal(t, BreakingEnumNoDelete, rules["ENUM_NO_DELETE"])
	assert.Equal(t, BreakingFieldNoDelete, rules["FIELD_NO_DELETE"])
	assert.Equal(t, BreakingRPCNoDelete, rules["RPC_NO_DELETE"])

	// Verify count is reasonable (we defined 59 breaking rules)
	assert.GreaterOrEqual(t, len(rules), 50)
}

func TestRuleCodeConsistency(t *testing.T) {
	t.Parallel()
	// Ensure all lint rules have LR category
	for ruleID, code := range AllLintRules() {
		assert.Equal(t, "LR", code.Category(), "lint rule %s should have LR category", ruleID)
	}

	// Ensure all breaking rules have BR category
	for ruleID, code := range AllBreakingRules() {
		assert.Equal(t, "BR", code.Category(), "breaking rule %s should have BR category", ruleID)
	}
}
