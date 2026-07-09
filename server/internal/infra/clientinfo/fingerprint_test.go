package clientinfo

import "testing"

func TestClientFingerprintStable(t *testing.T) {
	a := ClientFingerprint("1.2.3.4", "Mozilla/5.0")
	b := ClientFingerprint("1.2.3.4", "Mozilla/5.0")
	if a == "" || a != b {
		t.Fatalf("expected stable fingerprint, got %q and %q", a, b)
	}
	if len(a) != 32 {
		t.Fatalf("expected 16-byte hex (32 chars), got len=%d value=%q", len(a), a)
	}
}

func TestClientFingerprintDiffersByIPOrUA(t *testing.T) {
	base := ClientFingerprint("1.2.3.4", "Mozilla/5.0")
	otherIP := ClientFingerprint("1.2.3.5", "Mozilla/5.0")
	otherUA := ClientFingerprint("1.2.3.4", "curl/8.0")
	if base == otherIP || base == otherUA {
		t.Fatalf("fingerprints should differ: base=%q ip=%q ua=%q", base, otherIP, otherUA)
	}
}

func TestClientFingerprintEmptyFallback(t *testing.T) {
	fp := ClientFingerprint("", "")
	if fp == "" {
		t.Fatal("empty ip/ua should still produce a fingerprint")
	}
}
