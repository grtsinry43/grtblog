package telemetry

import "testing"

func TestValidateEndpointURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		// Valid
		{"valid https", "https://telemetry.example.com/collect", false},
		{"valid https with port", "https://telemetry.example.com:8443/collect", false},

		// Scheme
		{"reject http", "http://telemetry.example.com/collect", true},
		{"reject empty scheme", "://example.com", true},
		{"reject file scheme", "file:///etc/passwd", true},
		{"reject ftp", "ftp://example.com", true},

		// Loopback / private
		{"reject localhost", "https://localhost/collect", true},
		{"reject 127.0.0.1", "https://127.0.0.1/collect", true},
		{"reject 10.x", "https://10.0.0.1/collect", true},
		{"reject 192.168.x", "https://192.168.1.1/collect", true},
		{"reject 172.16.x", "https://172.16.0.1/collect", true},
		{"reject ::1", "https://[::1]/collect", true},

		// Link-local
		{"reject 169.254.x", "https://169.254.169.254/latest", true},

		// Cloud metadata
		{"reject metadata.google.internal", "https://metadata.google.internal/computeMetadata", true},

		// Empty / malformed
		{"reject empty", "", true},
		{"reject no host", "https:///path", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateEndpointURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateEndpointURL(%q) error = %v, wantErr = %v", tt.url, err, tt.wantErr)
			}
		})
	}
}
