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
