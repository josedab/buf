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
	"context"
	"sort"

	"github.com/bufbuild/buf/private/bufpkg/bufanalysis"
	"github.com/bufbuild/buf/private/bufpkg/bufprotosource"
)

// Checker runs custom DSL lint rules against protobuf files.
type Checker struct {
	config *CustomRulesConfig
}

// NewChecker creates a new DSL rule checker.
func NewChecker(config *CustomRulesConfig) *Checker {
	return &Checker{
		config: config,
	}
}

// Check runs all custom DSL rules against the given files and returns any violations.
func (c *Checker) Check(ctx context.Context, files []bufprotosource.File) ([]bufanalysis.FileAnnotation, error) {
	if c.config == nil || len(c.config.Rules) == 0 {
		return nil, nil
	}

	var allAnnotations []bufanalysis.FileAnnotation

	for _, file := range files {
		// Skip imports
		if file.IsImport() {
			continue
		}

		for _, rule := range c.config.Rules {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
			}

			annotations, err := rule.Check(ctx, file)
			if err != nil {
				return nil, err
			}
			allAnnotations = append(allAnnotations, annotations...)
		}
	}

	// Sort annotations by file path, line, column
	sort.Slice(allAnnotations, func(i, j int) bool {
		ai, aj := allAnnotations[i], allAnnotations[j]

		// Compare by file path
		pathI := ""
		pathJ := ""
		if ai.FileInfo() != nil {
			pathI = ai.FileInfo().Path()
		}
		if aj.FileInfo() != nil {
			pathJ = aj.FileInfo().Path()
		}
		if pathI != pathJ {
			return pathI < pathJ
		}

		// Compare by start line
		if ai.StartLine() != aj.StartLine() {
			return ai.StartLine() < aj.StartLine()
		}

		// Compare by start column
		if ai.StartColumn() != aj.StartColumn() {
			return ai.StartColumn() < aj.StartColumn()
		}

		// Compare by type (rule ID)
		return ai.Type() < aj.Type()
	})

	return allAnnotations, nil
}

// CheckFile runs all custom DSL rules against a single file.
func (c *Checker) CheckFile(ctx context.Context, file bufprotosource.File) ([]bufanalysis.FileAnnotation, error) {
	return c.Check(ctx, []bufprotosource.File{file})
}

// RuleCount returns the number of custom rules configured.
func (c *Checker) RuleCount() int {
	if c.config == nil {
		return 0
	}
	return len(c.config.Rules)
}

// RuleIDs returns the IDs of all configured custom rules.
func (c *Checker) RuleIDs() []string {
	if c.config == nil {
		return nil
	}
	return c.config.RuleIDs()
}
