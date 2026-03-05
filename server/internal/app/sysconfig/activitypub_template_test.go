package sysconfig

import "testing"

func TestValidateActivityPubPublishTemplate(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		wantErr bool
	}{
		{name: "empty allowed", raw: "", wantErr: false},
		{name: "supported fields", raw: "{{ .Title }} {{ .Summary }} {{ .URL }} {{ .ContentType }}", wantErr: false},
		{name: "unsupported field", raw: "{{ .Foo }}", wantErr: true},
		{name: "invalid syntax", raw: "{{ if .Title }", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateActivityPubPublishTemplate(tt.raw)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateActivityPubPublishTemplate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
