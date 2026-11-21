package app

import (
	"crypto/sha1"
	"crypto/sha256"
	"crypto/x509"
	"fmt"
	"strings"
)

func certificateOf(c *x509.Certificate) Certificate {
	s1 := sha1.Sum(c.Raw)
	s256 := sha256.Sum256(c.Raw)
	return Certificate{
		Subject:   c.Subject.String(),
		Issuer:    c.Issuer.String(),
		Serial:    fmt.Sprintf("%x", c.SerialNumber),
		NotBefore: c.NotBefore,
		NotAfter:  c.NotAfter,
		SHA1:      fingerprint(s1[:]),
		SHA256:    fingerprint(s256[:]),
		Debug:     c.Subject.CommonName == "Android Debug",
	}
}

func fingerprint(sum []byte) string {
	parts := make([]string, len(sum))
	for i, b := range sum {
		parts[i] = fmt.Sprintf("%02X", b)
	}
	return strings.Join(parts, ":")
}
