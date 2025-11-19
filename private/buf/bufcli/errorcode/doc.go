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
Package errorcode provides structured error codes for buf CLI errors.

# Error Code Format

Error codes follow the format BUF{category}{number} where:
  - BUF is the prefix for all buf errors
  - category is a two-letter category code
  - number is a three-digit error number

Example: BUFAU001 for Authentication: Invalid token

# Error Categories

  - AU: Authentication - Token, login, credential errors
  - MD: Module - Module resolution, dependency errors
  - WS: Workspace - Workspace configuration errors
  - CF: Config - buf.yaml, buf.gen.yaml errors
  - LR: Lint Rule - Individual lint rule violations
  - BR: Breaking Rule - Individual breaking rule violations
  - GN: Generate - Code generation errors
  - NW: Network - Connection, TLS, DNS errors
  - FS: Filesystem - File access, permission errors
  - IN: Internal - Bug/system errors

# Usage

Creating a new structured error:

	err := errorcode.NewError(
		errorcode.ModuleNotFound,
		"module \"buf.build/myorg/mymodule\" not found",
		errorcode.WithHelp("Run 'buf registry login' if this is a private module"),
	)

Using helper functions:

	err := errorcode.NewModuleNotFoundError("buf.build/myorg/mymodule")

Wrapping an existing error:

	err := errorcode.Wrap(errorcode.NetworkDNSError, originalErr)

Checking error codes:

	if errorcode.HasCode(err, errorcode.ModuleNotFound) {
		// Handle module not found
	}

Getting the error code from any error:

	code := errorcode.GetCode(err)
	if code != "" {
		fmt.Printf("Error code: %s\n", code)
	}

Formatting errors for display:

	err := errorcode.NewModuleNotFoundError("buf.build/myorg/mymodule")
	fmt.Print(err.Format())
	// Output:
	// BUFMD001: module "buf.build/myorg/mymodule" not found
	//   → Ensure the module exists on the registry or check your dependencies.
	//   → Run 'buf registry login' if this is a private module.
	//   → See: https://buf.build/docs/errors/BUFMD001

# Mapping Lint and Breaking Rules

Lint and breaking rule IDs can be converted to error codes:

	code := errorcode.LintRuleToCode("FIELD_LOWER_SNAKE_CASE")
	// Returns LintFieldLowerSnakeCase (BUFLR015)

	code := errorcode.BreakingRuleToCode("ENUM_NO_DELETE")
	// Returns BreakingEnumNoDelete (BUFBR001)

	// Or use RuleToCode for any rule type
	code := errorcode.RuleToCode("MESSAGE_PASCAL_CASE")
*/
package errorcode
