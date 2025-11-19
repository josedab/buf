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

/*
Package buflintdsl provides a domain-specific language for defining custom lint rules
without writing Go code. This makes rule authoring accessible to all teams.

# Overview

The DSL allows you to define custom lint rules in YAML format that can be included
in your buf.yaml configuration. Rules are defined using CEL (Common Expression Language)
for condition matching.

# Rule Structure

Each custom rule has the following structure:

	custom_rules:
	  - id: RULE_ID
	    message: "Error message template"
	    severity: error|warning
	    match:
	      kind: message|field|enum|enum_value|service|rpc|file|oneof
	      where:
	        - expression1
	        - expression2

# Match Kinds

The following match kinds are supported:

  - message: Match message descriptors
  - field: Match field descriptors
  - enum: Match enum descriptors
  - enum_value: Match enum value descriptors
  - service: Match service descriptors
  - rpc: Match RPC/method descriptors
  - file: Match file descriptors
  - oneof: Match oneof descriptors

# Context Variables

Each match kind provides different context variables for use in expressions:

Message context:

	name, full_name, nested_name, comment, deprecated, is_map_entry,
	field_count, oneof_count, nested_message_count, nested_enum_count, has_comment

Field context:

	name, full_name, number, field_type, type_name, label, comment, deprecated,
	is_repeated, is_map, is_optional, is_required, has_comment, json_name,
	message_name, has_default, in_oneof

Enum context:

	name, full_name, nested_name, comment, deprecated, value_count,
	has_comment, allow_alias

Enum value context:

	name, full_name, number, comment, deprecated, has_comment, enum_name

Service context:

	name, full_name, comment, deprecated, method_count, has_comment

RPC context:

	name, full_name, comment, deprecated, input_type, output_type,
	client_streaming, server_streaming, has_comment, service_name

File context:

	path, package, syntax, message_count, enum_count, service_count, deprecated

Oneof context:

	name, full_name, comment, has_comment, field_count, message_name

# Built-in Functions

String functions (CEL standard library - use method syntax):

	str.startsWith(prefix) - Check if string starts with prefix
	str.endsWith(suffix) - Check if string ends with suffix
	str.matches(regex) - Check if string matches regex pattern
	str.contains(substr) - Check if string contains substring

Case conversion:

	lower_snake_case(str) - Convert to lower_snake_case
	upper_snake_case(str) - Convert to UPPER_SNAKE_CASE
	pascal_case(str) - Convert to PascalCase
	camel_case(str) - Convert to camelCase
	to_lower(str) - Convert to lowercase
	to_upper(str) - Convert to uppercase

# Message Templates

Messages support template variables that are replaced with actual values:

	{{message.name}} - Message name
	{{message.full_name}} - Full message name
	{{field.name}} - Field name
	{{field.full_name}} - Full field name
	{{enum.name}} - Enum name
	{{value.name}} - Enum value name
	{{service.name}} - Service name
	{{rpc.name}} - RPC method name
	{{file.path}} - File path
	{{file.package}} - Package name
	{{oneof.name}} - Oneof name
	{{len(fields)}} - Number of fields

# Example Rules

Require comments on public fields:

	- id: PUBLIC_FIELD_COMMENT
	  message: "Public field '{{field.name}}' must have a comment"
	  match:
	    kind: field
	    where:
	      - has_comment == false
	      - !name.startsWith("_")

Enforce RPC naming convention:

	- id: RPC_NAMING
	  message: "RPC '{{rpc.name}}' must start with a verb"
	  match:
	    kind: rpc
	    where:
	      - !name.matches("^(Get|List|Create|Update|Delete|Watch|Search)")

Limit message complexity:

	- id: MESSAGE_COMPLEXITY
	  message: "Message '{{message.name}}' has {{len(fields)}} fields (max 30)"
	  match:
	    kind: message
	    where:
	      - field_count > 30

Require enum value prefix:

	- id: ENUM_VALUE_PREFIX
	  message: "Enum value must be prefixed with enum name"
	  match:
	    kind: enum_value
	    where:
	      - !name.startsWith(upper_snake_case(enum_name) + "_")

# Testing Rules

Rules can include inline tests:

	- id: TEST_RULE
	  message: "Test message"
	  match:
	    kind: message
	    where:
	      - field_count > 10
	  tests:
	    - input: |
	        message Small {
	          string a = 1;
	        }
	      expect_violation: false
	    - input: |
	        message Large {
	          string a = 1;
	          string b = 2;
	          // ... more fields
	        }
	      expect_violation: true

# Usage

To use custom lint rules, add them to your buf.yaml:

	version: v2
	lint:
	  use:
	    - STANDARD
	  custom_rules:
	    - id: MY_CUSTOM_RULE
	      message: "Custom rule violation"
	      match:
	        kind: message
	        where:
	          - field_count > 20

Then run buf lint as usual. Custom rule violations will be reported alongside
built-in rule violations.
*/
package buflintdsl
