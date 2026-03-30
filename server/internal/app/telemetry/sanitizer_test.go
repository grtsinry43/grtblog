package telemetry

import (
	"strings"
	"testing"
)

func TestSanitiseMessage(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "uuid replaced",
			input: "article not found: 550e8400-e29b-41d4-a716-446655440000",
			want:  "article not found: {uuid}",
		},
		{
			name:  "email replaced",
			input: "failed to send to user@example.com",
			want:  "failed to send to {email}",
		},
		{
			name:  "ipv4 replaced",
			input: "connection refused from 192.168.1.42",
			want:  "connection refused from {ip}",
		},
		{
			name:  "jwt replaced",
			input: "token expired: eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U",
			want:  "token expired: {jwt}",
		},
		{
			name:  "no pii untouched",
			input: "template execution failed: missing closing tag",
			want:  "template execution failed: missing closing tag",
		},
		{
			name:  "small numbers preserved",
			input: "connection pool exhausted after 30s timeout",
			want:  "connection pool exhausted after 30s timeout",
		},
		{
			name:  "large numeric id replaced",
			input: "record 1234567 not found in table",
			want:  "record {id} not found in table",
		},
		{
			name:  "ipv6 replaced",
			input: "connection from 2001:0db8:85a3:0000:0000:8a2e:0370:7334",
			want:  "connection from {ipv6}",
		},
		{
			name:  "short hex not treated as ipv6",
			input: "GORM error code: EAF123",
			want:  "GORM error code: EAF123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitiseMessage(tt.input)
			if got != tt.want {
				t.Errorf("SanitiseMessage(%q)\n  got  = %q\n  want = %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestNormaliseStack(t *testing.T) {
	raw := []byte(`goroutine 1 [running]:
runtime/debug.Stack()
	/usr/local/go/src/runtime/debug/stack.go:24 +0x5e
github.com/grtsinry43/grtblog-v2/server/internal/app/content.(*Service).GenerateHTML(0xc0001a2340, {0xc0004e2000, 0x42})
	/home/deploy/server/internal/app/content/service.go:142 +0x3a4
github.com/grtsinry43/grtblog-v2/server/internal/http/handler.(*ArticleHandler).Refresh(0xc0001b0000, 0xc000512000)
	/home/deploy/server/internal/http/handler/article_handler.go:88 +0x1f2
`)

	got := NormaliseStack(raw)

	if got == "" {
		t.Fatal("NormaliseStack returned empty string")
	}
	if strings.Contains(got, "/home/deploy") {
		t.Errorf("NormaliseStack should strip absolute paths, got:\n%s", got)
	}
	if strings.Contains(got, ":142") || strings.Contains(got, ":88") {
		t.Errorf("NormaliseStack should strip line numbers, got:\n%s", got)
	}
	if !strings.Contains(got, "internal/app/content") {
		t.Errorf("NormaliseStack should contain package path, got:\n%s", got)
	}
}

func TestNormaliseStack_EdgeCases(t *testing.T) {
	// Standard library function with no receiver.
	raw := []byte(`goroutine 1 [running]:
runtime.gopanic({0x1234, 0x5678})
	/usr/local/go/src/runtime/panic.go:1234 +0x100
main.main()
	/home/user/project/main.go:10 +0x20
`)
	got := NormaliseStack(raw)
	if got == "" {
		t.Fatal("NormaliseStack returned empty for stdlib stack")
	}
	if strings.Contains(got, "/usr/local") {
		t.Errorf("should strip stdlib paths, got:\n%s", got)
	}
}

func TestFingerprint_Deterministic(t *testing.T) {
	fp1 := Fingerprint("SERVER_ERROR", "internal/app/content.(*Service).GenerateHTML")
	fp2 := Fingerprint("SERVER_ERROR", "internal/app/content.(*Service).GenerateHTML")
	if fp1 != fp2 {
		t.Errorf("Fingerprint is not deterministic: %q != %q", fp1, fp2)
	}

	fp3 := Fingerprint("NOT_FOUND", "internal/app/content.(*Service).GenerateHTML")
	if fp1 == fp3 {
		t.Error("different biz codes should produce different fingerprints")
	}
}
