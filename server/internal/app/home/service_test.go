package home

import "testing"

func TestFirstCommaSeparatedImageURL(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{name: "single image", raw: "https://example.com/one.jpg", want: "https://example.com/one.jpg"},
		{name: "multiple images", raw: " https://example.com/one.jpg, https://example.com/two.jpg ", want: "https://example.com/one.jpg"},
		{name: "empty image", raw: "", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := firstCommaSeparatedImageURL(tt.raw); got != tt.want {
				t.Fatalf("firstCommaSeparatedImageURL(%q) = %q, want %q", tt.raw, got, tt.want)
			}
		})
	}
}
