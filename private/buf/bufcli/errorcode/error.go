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
	"encoding/json"
	"fmt"
	"strings"
)

// StructuredError represents a structured error with a code, message, and optional details.
type StructuredError struct {
	// Code is the structured error code.
	Code Code `json:"code"`
	// Message is the primary error message.
	Message string `json:"message"`
	// Detail provides additional context about the error.
	Detail string `json:"detail,omitempty"`
	// Help provides actionable guidance for resolving the error.
	Help string `json:"help,omitempty"`
	// underlying is the underlying error if wrapping another error.
	underlying error
}

// Error implements the error interface.
func (e *StructuredError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying error.
func (e *StructuredError) Unwrap() error {
	return e.underlying
}

// Format formats the error for display.
func (e *StructuredError) Format() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("%s: %s\n", e.Code, e.Message))
	if e.Detail != "" {
		b.WriteString(fmt.Sprintf("  %s\n", e.Detail))
	}
	if e.Help != "" {
		for _, line := range strings.Split(e.Help, "\n") {
			b.WriteString(fmt.Sprintf("  → %s\n", line))
		}
	}
	b.WriteString(fmt.Sprintf("  → See: %s\n", e.Code.URL()))
	return b.String()
}

// MarshalJSON implements json.Marshaler.
func (e *StructuredError) MarshalJSON() ([]byte, error) {
	type jsonError struct {
		Code    Code   `json:"code"`
		Message string `json:"message"`
		Detail  string `json:"detail,omitempty"`
		Help    string `json:"help,omitempty"`
		URL     string `json:"url"`
	}
	return json.Marshal(jsonError{
		Code:    e.Code,
		Message: e.Message,
		Detail:  e.Detail,
		Help:    e.Help,
		URL:     e.Code.URL(),
	})
}

// ErrorOption is a functional option for configuring a StructuredError.
type ErrorOption func(*StructuredError)

// WithDetail adds detail to the error.
func WithDetail(detail string) ErrorOption {
	return func(e *StructuredError) {
		e.Detail = detail
	}
}

// WithDetailf adds formatted detail to the error.
func WithDetailf(format string, args ...interface{}) ErrorOption {
	return func(e *StructuredError) {
		e.Detail = fmt.Sprintf(format, args...)
	}
}

// WithHelp adds help text to the error.
func WithHelp(help string) ErrorOption {
	return func(e *StructuredError) {
		e.Help = help
	}
}

// WithHelpf adds formatted help text to the error.
func WithHelpf(format string, args ...interface{}) ErrorOption {
	return func(e *StructuredError) {
		e.Help = fmt.Sprintf(format, args...)
	}
}

// WithUnderlying sets the underlying error.
func WithUnderlying(err error) ErrorOption {
	return func(e *StructuredError) {
		e.underlying = err
	}
}

// NewError creates a new StructuredError with the given code and message.
func NewError(code Code, message string, options ...ErrorOption) *StructuredError {
	err := &StructuredError{
		Code:    code,
		Message: message,
	}
	for _, opt := range options {
		opt(err)
	}
	return err
}

// NewErrorf creates a new StructuredError with a formatted message.
func NewErrorf(code Code, format string, args ...interface{}) *StructuredError {
	return &StructuredError{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
	}
}

// Wrap wraps an existing error with a structured error code.
func Wrap(code Code, err error, options ...ErrorOption) *StructuredError {
	if err == nil {
		return nil
	}
	structuredErr := &StructuredError{
		Code:       code,
		Message:    err.Error(),
		underlying: err,
	}
	for _, opt := range options {
		opt(structuredErr)
	}
	return structuredErr
}

// IsStructuredError returns true if the error is a StructuredError.
func IsStructuredError(err error) bool {
	_, ok := err.(*StructuredError)
	return ok
}

// GetCode returns the error code if the error is a StructuredError, or empty string otherwise.
func GetCode(err error) Code {
	if structuredErr, ok := err.(*StructuredError); ok {
		return structuredErr.Code
	}
	return ""
}

// HasCode returns true if the error has the given code.
func HasCode(err error, code Code) bool {
	return GetCode(err) == code
}
