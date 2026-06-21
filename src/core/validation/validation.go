// Package validation registers custom request validators used across features.
//
// The primary goal is a backend anti-XSS defense: even though output encoding is
// ultimately the frontend's responsibility, we reject obviously dangerous markup
// at the input boundary so script payloads never get persisted in the first place.
package validation

import (
	"html"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// htmlTagPattern matches anything that looks like an HTML/XML tag, e.g. <script>,
// </a>, <img src=...>. This is a deliberately broad heuristic for user-facing
// free-text fields that should never contain markup.
var htmlTagPattern = regexp.MustCompile(`<[^>]*>`)

// RegisterCustomValidators wires the project's custom validators into Gin's
// default validator engine. It must be called once at startup, before routes
// handle requests. It is safe to call multiple times.
func RegisterCustomValidators() error {
	v, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		// Gin is not using go-playground/validator; nothing to register.
		return nil
	}

	// "nohtml" fails when the field contains HTML tags or known XSS markers.
	if err := v.RegisterValidation("nohtml", validateNoHTML); err != nil {
		return err
	}

	return nil
}

// validateNoHTML returns false (invalid) if the string contains HTML tags or
// common script-injection markers.
func validateNoHTML(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	return !ContainsHTML(value)
}

// ContainsHTML reports whether s contains HTML tags or common XSS vectors.
func ContainsHTML(s string) bool {
	if htmlTagPattern.MatchString(s) {
		return true
	}
	lower := strings.ToLower(s)
	// Catch attribute-style and protocol-style vectors that may not be wrapped
	// in angle brackets (e.g. "javascript:alert(1)", "onerror=...").
	for _, marker := range []string{"javascript:", "data:text/html", "onerror=", "onload=", "<script", "&#"} {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return false
}

// SanitizeString escapes HTML-significant characters in s. Use this as a
// defense-in-depth measure for stored free text that legitimately may contain
// characters like < or & but must never be interpreted as markup.
func SanitizeString(s string) string {
	return html.EscapeString(strings.TrimSpace(s))
}
