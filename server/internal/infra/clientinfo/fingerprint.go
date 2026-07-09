package clientinfo

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// ClientFingerprint returns a stable short hash of IP + User-Agent.
// Used for like/view anti-abuse dedupe when client-supplied visitorId is untrusted.
func ClientFingerprint(ip, userAgent string) string {
	raw := strings.TrimSpace(ip) + "|" + strings.TrimSpace(userAgent)
	if raw == "|" {
		raw = "unknown"
	}
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:16])
}
