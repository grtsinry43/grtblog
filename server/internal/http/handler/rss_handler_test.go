package handler

import "testing"

func TestIsPrefetchPurpose(t *testing.T) {
	cases := []struct {
		name   string
		value  string
		expect bool
	}{
		{name: "empty", value: "", expect: false},
		{name: "normal", value: "navigate", expect: false},
		{name: "purpose_prefetch", value: "prefetch", expect: true},
		{name: "sec_purpose_prefetch_with_extra", value: "prefetch;prerender", expect: true},
		{name: "case_insensitive", value: "PreFetch", expect: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := isPrefetchPurpose(tc.value)
			if got != tc.expect {
				t.Fatalf("isPrefetchPurpose(%q)=%v, expect=%v", tc.value, got, tc.expect)
			}
		})
	}
}

func TestShouldRecordRSSAccessHeaders(t *testing.T) {
	base := map[string]string{
		"Purpose":     "",
		"Sec-Purpose": "",
		"X-Purpose":   "",
		"X-Moz":       "",
	}

	getHeader := func(headers map[string]string) func(string, ...string) string {
		return func(key string, defaultValue ...string) string {
			if v, ok := headers[key]; ok {
				return v
			}
			if len(defaultValue) > 0 {
				return defaultValue[0]
			}
			return ""
		}
	}

	if !shouldRecordRSSAccessHeaders(getHeader(base)) {
		t.Fatalf("expected record=true when no prefetch headers")
	}

	prefetchHeaders := []struct {
		name string
		key  string
	}{
		{name: "purpose", key: "Purpose"},
		{name: "sec-purpose", key: "Sec-Purpose"},
		{name: "x-purpose", key: "X-Purpose"},
		{name: "x-moz", key: "X-Moz"},
	}

	for _, tc := range prefetchHeaders {
		t.Run(tc.name, func(t *testing.T) {
			headers := map[string]string{
				"Purpose":     "",
				"Sec-Purpose": "",
				"X-Purpose":   "",
				"X-Moz":       "",
			}
			headers[tc.key] = "prefetch"
			if shouldRecordRSSAccessHeaders(getHeader(headers)) {
				t.Fatalf("expected record=false when %s is prefetch", tc.key)
			}
		})
	}
}

func TestBuildRSSEnclosure(t *testing.T) {
	if enclosure := buildRSSEnclosure(""); enclosure != nil {
		t.Fatalf("expected nil enclosure for empty image url")
	}

	enclosure := buildRSSEnclosure("https://cdn.example.com/a.png?x=1")
	if enclosure == nil {
		t.Fatalf("expected enclosure")
	}
	if enclosure.Url != "https://cdn.example.com/a.png?x=1" {
		t.Fatalf("unexpected enclosure url: %q", enclosure.Url)
	}
	if enclosure.Length != "0" {
		t.Fatalf("unexpected enclosure length: %q", enclosure.Length)
	}
	if enclosure.Type != "image/png" {
		t.Fatalf("unexpected enclosure type: %q", enclosure.Type)
	}
}

func TestDetectImageMIMEType(t *testing.T) {
	cases := []struct {
		url    string
		expect string
	}{
		{url: "https://cdn.example.com/a.jpg", expect: "image/jpeg"},
		{url: "https://cdn.example.com/a.jpeg?x=1", expect: "image/jpeg"},
		{url: "https://cdn.example.com/a.png", expect: "image/png"},
		{url: "https://cdn.example.com/a.gif", expect: "image/gif"},
		{url: "https://cdn.example.com/a.webp", expect: "image/webp"},
		{url: "https://cdn.example.com/a.avif", expect: "image/avif"},
		{url: "https://cdn.example.com/a.svg", expect: "image/svg+xml"},
		{url: "https://cdn.example.com/a.bmp", expect: "image/bmp"},
		{url: "https://cdn.example.com/a.ico", expect: "image/x-icon"},
		{url: "https://cdn.example.com/image", expect: "image/jpeg"},
		{url: "", expect: "image/jpeg"},
	}

	for _, tc := range cases {
		if got := detectImageMIMEType(tc.url); got != tc.expect {
			t.Fatalf("detectImageMIMEType(%q)=%q, expect=%q", tc.url, got, tc.expect)
		}
	}
}
