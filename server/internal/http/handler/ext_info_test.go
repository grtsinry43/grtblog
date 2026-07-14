package handler

import (
	"strings"
	"testing"

	"github.com/grtsinry43/grtblog-v2/server/internal/http/contract"
)

func TestParseMomentExtInfo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		payload    string
		wantErr    string
		wantResult bool
	}{
		{
			name:       "accepts supported atmosphere and preserves other keys",
			payload:    `{"images":[{"id":"cover.webp"}],"moment":{"weather":"rainy","mood":"calm"},"custom":true}`,
			wantResult: true,
		},
		{
			name:       "accepts omitted atmosphere",
			payload:    `{"images":[]}`,
			wantResult: true,
		},
		{
			name:       "accepts nullable atmosphere values",
			payload:    `{"moment":{"weather":null,"mood":null}}`,
			wantResult: true,
		},
		{
			name:    "rejects non object moment metadata",
			payload: `{"moment":"sunny"}`,
			wantErr: "moment must be object",
		},
		{
			name:    "rejects non string weather",
			payload: `{"moment":{"weather":1}}`,
			wantErr: "moment.weather must be string",
		},
		{
			name:    "rejects unknown weather",
			payload: `{"moment":{"weather":"storm"}}`,
			wantErr: `moment.weather has unsupported value "storm"`,
		},
		{
			name:    "rejects unknown mood",
			payload: `{"moment":{"mood":"angry"}}`,
			wantErr: `moment.mood has unsupported value "angry"`,
		},
		{
			name:    "keeps generic image validation",
			payload: `{"images":[{}],"moment":{"weather":"sunny"}}`,
			wantErr: "images item missing id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			raw := contract.JSONRaw([]byte(tt.payload))
			got, err := parseMomentExtInfo(&raw)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("expected error containing %q, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("parse moment ext info: %v", err)
			}
			if tt.wantResult && string(got) != tt.payload {
				t.Fatalf("payload changed: got %s want %s", got, tt.payload)
			}
		})
	}
}
