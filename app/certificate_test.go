package app

import (
	"strings"
	"testing"
)

func TestCertificateOf(t *testing.T) {
	cert := testCertificate(t, "Android Debug")
	got := certificateOf(cert)
	if !got.Debug || !strings.Contains(got.Subject, "Android Debug") || got.Serial != "1267" {
		t.Errorf("certificate = %+v", got)
	}
	if !got.NotBefore.Equal(cert.NotBefore) || !got.NotAfter.Equal(cert.NotAfter) {
		t.Errorf("validity = %v - %v", got.NotBefore, got.NotAfter)
	}
	if len(got.SHA1) != 20*3-1 || len(got.SHA256) != 32*3-1 {
		t.Errorf("fingerprints = %s, %s", got.SHA1, got.SHA256)
	}
	if certificateOf(testCertificate(t, "Release")).Debug {
		t.Error("a release key is not debug")
	}
}

func TestFingerprint(t *testing.T) {
	if got := fingerprint([]byte{0x0a, 0xff}); got != "0A:FF" {
		t.Errorf("fingerprint = %s", got)
	}
}
