package telemetry

import (
	"crypto/sha256"
	"fmt"
	"regexp"
	"strings"
)

// sanitizer strips personally identifiable information (PII) and
// instance-specific details from error messages and stack traces,
// producing deterministic fingerprints suitable for anonymous reporting.

// --- message sanitisation -----------------------------------------------------------

// compiledRules is a pre-compiled slice of regex → replacement pairs.
var compiledRules []sanitiseRule

type sanitiseRule struct {
	re   *regexp.Regexp
	repl string
}

func init() {
	patterns := []struct {
		expr string
		repl string
	}{
		// JWT-like (three base64url segments separated by dots) — must run before hex/token rules
		{`eyJ[A-Za-z0-9_-]+\.eyJ[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+`, "{jwt}"},
		// UUIDs  (8-4-4-4-12)
		{`[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}`, "{uuid}"},
		// Email addresses — must run before generic ID
		{`[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`, "{email}"},
		// IPv4 — must run before generic numeric ID
		{`\b\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}\b`, "{ip}"},
		// IPv6 — require at least two colon-separated hex groups to avoid matching random hex strings
		{`(?:[0-9a-fA-F]{1,4}:){2,7}[0-9a-fA-F]{0,4}|::(?:[0-9a-fA-F]{1,4}:){0,5}[0-9a-fA-F]{0,4}`, "{ipv6}"},
		// URL path segments that look like slugs with IDs
		{`/[a-zA-Z0-9_-]+/[0-9a-fA-F]{6,}`, "/{resource}/{id}"},
		// Hex tokens (≥16 chars)
		{`\b[0-9a-fA-F]{16,}\b`, "{token}"},
		// File system absolute paths
		{`(?:/[a-zA-Z0-9._-]+){3,}`, "{path}"},
		// Windows absolute paths
		{`[A-Z]:\\(?:[a-zA-Z0-9._-]+\\){2,}`, "{path}"},
		// Long numeric IDs (≥6 digits) — only strip numbers that look like database IDs, not small numbers
		// like timeout values, port numbers, or error codes
		{`\b\d{6,}\b`, "{id}"},
	}

	compiledRules = make([]sanitiseRule, 0, len(patterns))
	for _, p := range patterns {
		compiledRules = append(compiledRules, sanitiseRule{
			re:   regexp.MustCompile(p.expr),
			repl: p.repl,
		})
	}
}

// SanitiseMessage removes PII from an error message string.
func SanitiseMessage(msg string) string {
	for _, rule := range compiledRules {
		msg = rule.re.ReplaceAllString(msg, rule.repl)
	}
	return msg
}

// --- stack trace normalisation ------------------------------------------------------

// NormaliseStack converts a raw runtime/debug.Stack() output to a
// compact, deterministic form: only package + function names, no
// absolute paths, no line numbers.
//
// Example input line:
//
//	github.com/grtsinry43/grtblog-v2/server/internal/app/content.(*Service).GenerateHTML(...)
//	    /home/deploy/server/internal/app/content/service.go:142 +0x3a4
//
// Example output line:
//
//	internal/app/content.(*Service).GenerateHTML
func NormaliseStack(raw []byte) string {
	lines := strings.Split(string(raw), "\n")
	var normalised []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Skip goroutine header ("goroutine 1 [running]:")
		if strings.HasPrefix(line, "goroutine ") {
			continue
		}
		// Skip file:line references (indented lines starting with / or drive letter)
		if strings.HasPrefix(line, "/") || (len(line) > 2 && line[1] == ':' && line[2] == '\\') {
			continue
		}
		// Skip lines that are purely hex offsets
		if strings.HasPrefix(line, "+0x") || strings.HasPrefix(line, "0x") {
			continue
		}

		// Extract package.Function part; strip argument list
		fn := line
		if idx := strings.Index(fn, "("); idx > 0 {
			// Keep receiver if present, e.g. "(*Service).Foo"
			fn = stripArgs(fn)
		}
		// Strip the module prefix to keep only internal paths
		fn = stripModulePrefix(fn)
		if fn != "" {
			normalised = append(normalised, fn)
		}
	}

	return strings.Join(normalised, "\n")
}

// stripModulePrefix removes the Go module path, keeping only the package-relative path.
// e.g. "github.com/grtsinry43/grtblog-v2/server/internal/app/foo.Bar" → "internal/app/foo.Bar"
func stripModulePrefix(s string) string {
	const marker = "/internal/"
	if idx := strings.Index(s, marker); idx >= 0 {
		return s[idx+1:] // "internal/app/foo.Bar"
	}
	// For standard library or third-party, keep as-is but trim module domain
	if idx := strings.LastIndex(s, "/vendor/"); idx >= 0 {
		return s[idx+8:]
	}
	return s
}

// stripArgs removes the trailing argument list from a function signature.
// "(*Service).Foo(0xc000123456, {0xc000...})" → "(*Service).Foo"
func stripArgs(s string) string {
	depth := 0
	for i := len(s) - 1; i >= 0; i-- {
		switch s[i] {
		case ')':
			depth++
		case '(':
			depth--
			if depth == 0 {
				// Check if this is a receiver like (*Service)
				if i > 0 && (s[i-1] == '.' || s[i-1] == '*') {
					continue
				}
				return strings.TrimSpace(s[:i])
			}
		}
	}
	return s
}

// --- fingerprinting -----------------------------------------------------------------

// Fingerprint produces a deterministic SHA-256 hex digest from a
// normalised stack trace + biz error code. Two identical bugs on
// different machines will produce the same fingerprint.
func Fingerprint(bizCode string, normalisedStack string) string {
	h := sha256.New()
	h.Write([]byte(bizCode))
	h.Write([]byte{0})
	h.Write([]byte(normalisedStack))
	return fmt.Sprintf("%x", h.Sum(nil))[:16] // 16-char prefix is sufficient
}
