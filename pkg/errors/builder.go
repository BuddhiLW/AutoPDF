// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package errors

// ErrorBuilder builds DomainError instances fluently
type ErrorBuilder struct {
	err *DomainError
}

func NewValidationError(code, message string) *ErrorBuilder {
	return &ErrorBuilder{err: &DomainError{Category: ErrValidation, Code: code, Message: message, Details: map[string]interface{}{}}}
}

func NewInternalError(code, message string) *ErrorBuilder {
	return &ErrorBuilder{err: &DomainError{Category: ErrInternal, Code: code, Message: message, Details: map[string]interface{}{}}}
}

func NewNotFoundError(code, message string) *ErrorBuilder {
	return &ErrorBuilder{err: &DomainError{Category: ErrNotFound, Code: code, Message: message, Details: map[string]interface{}{}}}
}

func NewInvalidInputError(code, message string) *ErrorBuilder {
	return &ErrorBuilder{err: &DomainError{Category: ErrInvalidInput, Code: code, Message: message, Details: map[string]interface{}{}}}
}

func (b *ErrorBuilder) WithBlame(blame string) *ErrorBuilder {
	b.err.Blame = blame
	return b
}

func (b *ErrorBuilder) WithDetails(details map[string]interface{}) *ErrorBuilder {
	if details == nil {
		return b
	}
	if b.err.Details == nil {
		b.err.Details = map[string]interface{}{}
	}
	for k, v := range details {
		b.err.Details[k] = v
	}
	return b
}

func (b *ErrorBuilder) WithDetail(key string, value interface{}) *ErrorBuilder {
	if b.err.Details == nil {
		b.err.Details = map[string]interface{}{}
	}
	b.err.Details[key] = value
	return b
}

func (b *ErrorBuilder) WithCause(cause error) *ErrorBuilder {
	b.err.Cause = cause
	return b
}

func (b *ErrorBuilder) WithSuggestions(suggestions ...string) *ErrorBuilder {
	b.err.Suggestions = append(b.err.Suggestions, suggestions...)
	return b
}

func (b *ErrorBuilder) Build() error { return b.err }
