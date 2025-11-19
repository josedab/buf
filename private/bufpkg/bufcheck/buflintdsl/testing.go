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
	"fmt"
	"strings"

	"github.com/bufbuild/buf/private/bufpkg/bufparse"
	"github.com/bufbuild/buf/private/bufpkg/bufprotosource"
	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/types/descriptorpb"
)

// RuleTest represents a test case for a DSL rule.
type RuleTest struct {
	// Name is an optional name for the test case.
	Name string
	// Input is the protobuf source content to test.
	Input string
	// ExpectViolation indicates whether a violation is expected.
	ExpectViolation bool
}

// TestResult represents the result of running a rule test.
type TestResult struct {
	// Test is the test that was run.
	Test RuleTest
	// Passed indicates whether the test passed.
	Passed bool
	// Error contains any error message if the test failed.
	Error string
	// Violations contains the actual violations found.
	Violations int
}

// TestRunner runs tests for DSL rules.
type TestRunner struct {
	rule *CompiledRule
}

// NewTestRunner creates a new test runner for a compiled rule.
func NewTestRunner(rule *CompiledRule) *TestRunner {
	return &TestRunner{
		rule: rule,
	}
}

// RunTest runs a single test case.
func (r *TestRunner) RunTest(ctx context.Context, test RuleTest) TestResult {
	result := TestResult{
		Test: test,
	}

	// Parse the input protobuf
	file, err := parseProtoContent(test.Input)
	if err != nil {
		result.Error = fmt.Sprintf("failed to parse input: %v", err)
		return result
	}

	// Run the rule
	annotations, err := r.rule.Check(ctx, file)
	if err != nil {
		result.Error = fmt.Sprintf("failed to check rule: %v", err)
		return result
	}

	result.Violations = len(annotations)

	// Check if result matches expectation
	if test.ExpectViolation && len(annotations) == 0 {
		result.Error = "expected violation but none found"
		return result
	}
	if !test.ExpectViolation && len(annotations) > 0 {
		var msgs []string
		for _, ann := range annotations {
			msgs = append(msgs, ann.Message())
		}
		result.Error = fmt.Sprintf("expected no violation but found %d: %s", len(annotations), strings.Join(msgs, "; "))
		return result
	}

	result.Passed = true
	return result
}

// RunTests runs multiple test cases.
func (r *TestRunner) RunTests(ctx context.Context, tests []RuleTest) []TestResult {
	results := make([]TestResult, len(tests))
	for i, test := range tests {
		results[i] = r.RunTest(ctx, test)
	}
	return results
}

// parseProtoContent parses protobuf content into a bufprotosource.File.
func parseProtoContent(content string) (bufprotosource.File, error) {
	// Wrap content in a valid proto file if it doesn't have syntax
	if !strings.Contains(content, "syntax") {
		content = "syntax = \"proto3\";\n\npackage test;\n\n" + content
	}

	// Create a file descriptor proto
	fdp := &descriptorpb.FileDescriptorProto{
		Name:    proto.String("test.proto"),
		Package: proto.String("test"),
	}

	// Parse the content (simplified parsing for testing)
	// In a real implementation, this would use a proper proto parser
	// For now, we'll create a minimal descriptor based on common patterns
	if err := parseIntoDescriptor(content, fdp); err != nil {
		return nil, err
	}

	// Build the file descriptor
	fdSet := &descriptorpb.FileDescriptorSet{
		File: []*descriptorpb.FileDescriptorProto{fdp},
	}

	files, err := protodesc.NewFiles(fdSet)
	if err != nil {
		return nil, fmt.Errorf("failed to create file registry: %w", err)
	}

	// Convert to bufprotosource.File
	inputFiles := []testInputFile{
		{
			fdp:  fdp,
			path: "test.proto",
		},
	}

	sourceFiles, err := bufprotosource.NewFiles(ctx, inputFiles, files)
	if err != nil {
		return nil, fmt.Errorf("failed to create source files: %w", err)
	}

	if len(sourceFiles) == 0 {
		return nil, fmt.Errorf("no files created")
	}

	return sourceFiles[0], nil
}

// parseIntoDescriptor parses proto content into a descriptor.
// This is a simplified parser for testing purposes.
func parseIntoDescriptor(content string, fdp *descriptorpb.FileDescriptorProto) error {
	lines := strings.Split(content, "\n")

	var currentMessage *descriptorpb.DescriptorProto
	var currentEnum *descriptorpb.EnumDescriptorProto
	var currentService *descriptorpb.ServiceDescriptorProto
	var fieldNumber int32 = 1
	var enumNumber int32 = 0

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}

		// Parse syntax
		if strings.HasPrefix(line, "syntax") {
			if strings.Contains(line, "proto3") {
				fdp.Syntax = proto.String("proto3")
			} else if strings.Contains(line, "proto2") {
				fdp.Syntax = proto.String("proto2")
			}
			continue
		}

		// Parse package
		if strings.HasPrefix(line, "package") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				pkg := strings.TrimSuffix(parts[1], ";")
				fdp.Package = proto.String(pkg)
			}
			continue
		}

		// Parse message
		if strings.HasPrefix(line, "message ") {
			name := strings.TrimPrefix(line, "message ")
			name = strings.TrimSuffix(name, " {")
			name = strings.TrimSuffix(name, "{")
			name = strings.TrimSpace(name)

			currentMessage = &descriptorpb.DescriptorProto{
				Name: proto.String(name),
			}
			currentEnum = nil
			currentService = nil
			fieldNumber = 1
			continue
		}

		// Parse enum
		if strings.HasPrefix(line, "enum ") {
			name := strings.TrimPrefix(line, "enum ")
			name = strings.TrimSuffix(name, " {")
			name = strings.TrimSuffix(name, "{")
			name = strings.TrimSpace(name)

			currentEnum = &descriptorpb.EnumDescriptorProto{
				Name: proto.String(name),
			}
			currentMessage = nil
			currentService = nil
			enumNumber = 0
			continue
		}

		// Parse service
		if strings.HasPrefix(line, "service ") {
			name := strings.TrimPrefix(line, "service ")
			name = strings.TrimSuffix(name, " {")
			name = strings.TrimSuffix(name, "{")
			name = strings.TrimSpace(name)

			currentService = &descriptorpb.ServiceDescriptorProto{
				Name: proto.String(name),
			}
			currentMessage = nil
			currentEnum = nil
			continue
		}

		// Handle closing braces
		if line == "}" {
			if currentMessage != nil {
				fdp.MessageType = append(fdp.MessageType, currentMessage)
				currentMessage = nil
			}
			if currentEnum != nil {
				fdp.EnumType = append(fdp.EnumType, currentEnum)
				currentEnum = nil
			}
			if currentService != nil {
				fdp.Service = append(fdp.Service, currentService)
				currentService = nil
			}
			continue
		}

		// Parse fields within a message
		if currentMessage != nil && !strings.HasPrefix(line, "//") {
			field := parseField(line, fieldNumber)
			if field != nil {
				currentMessage.Field = append(currentMessage.Field, field)
				fieldNumber++
			}
		}

		// Parse enum values
		if currentEnum != nil && !strings.HasPrefix(line, "//") {
			value := parseEnumValue(line, enumNumber)
			if value != nil {
				currentEnum.Value = append(currentEnum.Value, value)
				enumNumber++
			}
		}

		// Parse RPC methods
		if currentService != nil && strings.HasPrefix(line, "rpc ") {
			method := parseMethod(line)
			if method != nil {
				currentService.Method = append(currentService.Method, method)
			}
		}
	}

	return nil
}

// parseField parses a field definition.
func parseField(line string, number int32) *descriptorpb.FieldDescriptorProto {
	// Handle simple field patterns like "string name = 1;"
	parts := strings.Fields(line)
	if len(parts) < 3 {
		return nil
	}

	fieldType := parts[0]
	fieldName := parts[1]

	// Remove trailing semicolon and equals
	fieldName = strings.TrimSuffix(fieldName, ";")
	if strings.Contains(fieldName, "=") {
		fieldName = strings.Split(fieldName, "=")[0]
	}
	fieldName = strings.TrimSpace(fieldName)

	var typ descriptorpb.FieldDescriptorProto_Type
	switch fieldType {
	case "string":
		typ = descriptorpb.FieldDescriptorProto_TYPE_STRING
	case "int32":
		typ = descriptorpb.FieldDescriptorProto_TYPE_INT32
	case "int64":
		typ = descriptorpb.FieldDescriptorProto_TYPE_INT64
	case "bool":
		typ = descriptorpb.FieldDescriptorProto_TYPE_BOOL
	case "bytes":
		typ = descriptorpb.FieldDescriptorProto_TYPE_BYTES
	case "double":
		typ = descriptorpb.FieldDescriptorProto_TYPE_DOUBLE
	case "float":
		typ = descriptorpb.FieldDescriptorProto_TYPE_FLOAT
	case "repeated":
		// Handle repeated fields
		if len(parts) >= 4 {
			return &descriptorpb.FieldDescriptorProto{
				Name:   proto.String(parts[2]),
				Number: proto.Int32(number),
				Type:   descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
				Label:  descriptorpb.FieldDescriptorProto_LABEL_REPEATED.Enum(),
			}
		}
		return nil
	default:
		// Assume it's a message type
		typ = descriptorpb.FieldDescriptorProto_TYPE_MESSAGE
	}

	return &descriptorpb.FieldDescriptorProto{
		Name:   proto.String(fieldName),
		Number: proto.Int32(number),
		Type:   typ.Enum(),
		Label:  descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
	}
}

// parseEnumValue parses an enum value definition.
func parseEnumValue(line string, number int32) *descriptorpb.EnumValueDescriptorProto {
	// Handle patterns like "VALUE_NAME = 0;"
	parts := strings.Fields(line)
	if len(parts) < 1 {
		return nil
	}

	name := parts[0]
	name = strings.TrimSuffix(name, ";")
	name = strings.TrimSuffix(name, "=")
	name = strings.TrimSpace(name)

	if name == "" {
		return nil
	}

	return &descriptorpb.EnumValueDescriptorProto{
		Name:   proto.String(name),
		Number: proto.Int32(number),
	}
}

// parseMethod parses an RPC method definition.
func parseMethod(line string) *descriptorpb.MethodDescriptorProto {
	// Handle patterns like "rpc GetUser(GetUserRequest) returns (GetUserResponse);"
	line = strings.TrimPrefix(line, "rpc ")
	line = strings.TrimSuffix(line, ";")
	line = strings.TrimSuffix(line, "{}")
	line = strings.TrimSpace(line)

	// Extract method name
	parenIdx := strings.Index(line, "(")
	if parenIdx == -1 {
		return nil
	}
	name := strings.TrimSpace(line[:parenIdx])

	// Extract input type
	closeParen := strings.Index(line, ")")
	if closeParen == -1 {
		return nil
	}
	inputType := strings.TrimSpace(line[parenIdx+1 : closeParen])

	// Extract output type
	returnsIdx := strings.Index(line, "returns")
	if returnsIdx == -1 {
		return nil
	}
	outputPart := line[returnsIdx+7:]
	openParen := strings.Index(outputPart, "(")
	closeParen = strings.Index(outputPart, ")")
	if openParen == -1 || closeParen == -1 {
		return nil
	}
	outputType := strings.TrimSpace(outputPart[openParen+1 : closeParen])

	return &descriptorpb.MethodDescriptorProto{
		Name:       proto.String(name),
		InputType:  proto.String("." + inputType),
		OutputType: proto.String("." + outputType),
	}
}

// testInputFile implements bufprotosource.InputFile for testing.
type testInputFile struct {
	fdp  *descriptorpb.FileDescriptorProto
	path string
}

func (f testInputFile) Path() string {
	return f.path
}

func (f testInputFile) ExternalPath() string {
	return f.path
}

func (f testInputFile) FullName() bufparse.FullName {
	return nil
}

func (f testInputFile) CommitID() uuid.UUID {
	return uuid.UUID{}
}

func (f testInputFile) IsImport() bool {
	return false
}

func (f testInputFile) FileDescriptorProto() *descriptorpb.FileDescriptorProto {
	return f.fdp
}

func (f testInputFile) IsSyntaxUnspecified() bool {
	return f.fdp.GetSyntax() == ""
}

func (f testInputFile) UnusedDependencyIndexes() []int32 {
	return nil
}

var ctx = context.Background()
