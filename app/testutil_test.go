package app

import (
	"archive/zip"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// openZip writes the files into a zip and returns its path and its files. The
// entries named in broken use a compression method that cannot be opened.
func openZip(t *testing.T, files map[string][]byte, broken ...string) (string, []*zip.File) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.zip")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	w := zip.NewWriter(f)
	for name, data := range files {
		fw, err := w.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		_, _ = fw.Write(data)
	}
	for _, name := range broken {
		fw, err := w.CreateRaw(&zip.FileHeader{Name: name, Method: 99, CompressedSize64: 3, UncompressedSize64: 3})
		if err != nil {
			t.Fatal(err)
		}
		_, _ = fw.Write([]byte("abc"))
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	r, err := zip.OpenReader(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = r.Close() })
	return path, r.File
}

// zipFiles is openZip for tests that only need the entries; the files are
// empty strings of the given names
func zipFiles(t *testing.T, names ...string) []*zip.File {
	t.Helper()
	files := map[string][]byte{}
	for _, name := range names {
		files[name] = []byte(name)
	}
	_, entries := openZip(t, files)
	return entries
}

// testCertificate makes a self signed certificate with the common name
func testCertificate(t *testing.T, commonName string) *x509.Certificate {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	template := &x509.Certificate{
		SerialNumber: big.NewInt(4711),
		Subject:      pkix.Name{CommonName: commonName},
		NotBefore:    time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		NotAfter:     time.Date(2036, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatal(err)
	}
	return cert
}
