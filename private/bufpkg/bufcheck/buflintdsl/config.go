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
	"fmt"
)

// ExternalCustomRule represents the external YAML representation of a custom rule.
type ExternalCustomRule struct {
	ID       string              `json:"id" yaml:"id"`
	Message  string              `json:"message" yaml:"message"`
	Severity string              `json:"severity,omitempty" yaml:"severity,omitempty"`
	Match    ExternalMatch       `json:"match" yaml:"match"`
	Tests    []ExternalRuleTest  `json:"tests,omitempty" yaml:"tests,omitempty"`
}

// ExternalMatch represents the external YAML representation of a match clause.
type ExternalMatch struct {
	Kind  string   `json:"kind" yaml:"kind"`
	Where []string `json:"where,omitempty" yaml:"where,omitempty"`
}

// ExternalRuleTest represents a test case for a custom rule.
type ExternalRuleTest struct {
	Input           string `json:"input" yaml:"input"`
	ExpectViolation bool   `json:"expect_violation" yaml:"expect_violation"`
}

// ParseExternalCustomRules converts external custom rules to RuleSpecs.
func ParseExternalCustomRules(externalRules []ExternalCustomRule) ([]RuleSpec, error) {
	specs := make([]RuleSpec, 0, len(externalRules))
	seenIDs := make(map[string]struct{})

	for i, ext := range externalRules {
		if ext.ID == "" {
			return nil, fmt.Errorf("custom rule %d: id is required", i+1)
		}

		// Check for duplicate IDs
		if _, ok := seenIDs[ext.ID]; ok {
			return nil, fmt.Errorf("custom rule %d: duplicate id %q", i+1, ext.ID)
		}
		seenIDs[ext.ID] = struct{}{}

		if ext.Message == "" {
			return nil, fmt.Errorf("custom rule %s: message is required", ext.ID)
		}
		if ext.Match.Kind == "" {
			return nil, fmt.Errorf("custom rule %s: match.kind is required", ext.ID)
		}

		// Parse severity, default to error
		severity := SeverityError
		if ext.Severity != "" {
			switch ext.Severity {
			case "error":
				severity = SeverityError
			case "warning":
				severity = SeverityWarning
			default:
				return nil, fmt.Errorf("custom rule %s: invalid severity %q, must be 'error' or 'warning'", ext.ID, ext.Severity)
			}
		}

		spec := RuleSpec{
			ID:       ext.ID,
			Message:  ext.Message,
			Severity: severity,
			Match: Match{
				Kind:  MatchKind(ext.Match.Kind),
				Where: ext.Match.Where,
			},
		}

		specs = append(specs, spec)
	}

	return specs, nil
}

// CustomRulesConfig holds the configuration for custom DSL lint rules.
type CustomRulesConfig struct {
	// Rules are the compiled DSL rules.
	Rules []*CompiledRule
}

// NewCustomRulesConfig creates a new CustomRulesConfig from external rules.
func NewCustomRulesConfig(externalRules []ExternalCustomRule) (*CustomRulesConfig, error) {
	specs, err := ParseExternalCustomRules(externalRules)
	if err != nil {
		return nil, err
	}

	rules := make([]*CompiledRule, 0, len(specs))
	for _, spec := range specs {
		rule, err := CompileRule(spec)
		if err != nil {
			return nil, fmt.Errorf("failed to compile custom rule %s: %w", spec.ID, err)
		}
		rules = append(rules, rule)
	}

	return &CustomRulesConfig{
		Rules: rules,
	}, nil
}

// RuleIDs returns the IDs of all custom rules.
func (c *CustomRulesConfig) RuleIDs() []string {
	if c == nil {
		return nil
	}
	ids := make([]string, len(c.Rules))
	for i, rule := range c.Rules {
		ids[i] = rule.ID()
	}
	return ids
}
