package app

import (
	"archive/zip"
	"crypto/x509"
	"fmt"
	"io"
	"strings"

	"github.com/fullsailor/pkcs7"
)

const (
	apkSigV2  = 0x7109871a
	apkSigV3  = 0xf05368c0
	apkSigV31 = 0x1b93ad61
)

type signatureScheme struct {
	id   uint32
	name string
}

// signatureSchemes are the schemes of the APK Signing Block
func signatureSchemes() []signatureScheme {
	return []signatureScheme{{apkSigV2, "v2"}, {apkSigV3, "v3"}, {apkSigV31, "v3.1"}}
}

// signingCollector gathers the schemes and the signer certificates, each
// certificate once
type signingCollector struct {
	signing *Signing
	seen    map[string]bool
}

// readAPKSigning finds the signature schemes and signer certificates of an
// apk: the APK Signing Block for v2 and later, META-INF for v1
func readAPKSigning(r io.ReaderAt, size int64, files []*zip.File) (*Signing, error) {
	c := &signingCollector{signing: &Signing{}, seen: map[string]bool{}}
	c.addV1(files)
	blocks, err := apkSigningBlocks(r, size)
	if err != nil {
		return c.signing, err
	}
	return c.signing, c.addBlocks(blocks)
}

func (c *signingCollector) addCerts(certs []*x509.Certificate) {
	for _, x := range certs {
		cert := certificateOf(x)
		if !c.seen[cert.SHA256] {
			c.seen[cert.SHA256] = true
			c.signing.Certificates = append(c.signing.Certificates, cert)
		}
	}
}

// addV1 reads the signature files in META-INF
func (c *signingCollector) addV1(files []*zip.File) {
	for _, f := range files {
		if isV1SignatureFile(f.Name) {
			c.addV1File(f)
		}
	}
}

func isV1SignatureFile(name string) bool {
	if !strings.HasPrefix(name, "META-INF/") || strings.Count(name, "/") != 1 {
		return false
	}
	switch strings.ToUpper(name[strings.LastIndex(name, ".")+1:]) {
	case "RSA", "DSA", "EC":
		return true
	}
	return false
}

func (c *signingCollector) addV1File(f *zip.File) {
	data, err := readZipFile(f, 1<<20)
	if err != nil {
		return
	}
	p7, err := pkcs7.Parse(data)
	if err != nil {
		return
	}
	if len(c.signing.Schemes) == 0 {
		c.signing.Schemes = append(c.signing.Schemes, "v1")
	}
	c.addCerts(p7.Certificates)
}

// addBlocks reads the schemes found in the APK Signing Block
func (c *signingCollector) addBlocks(blocks map[uint32][]byte) error {
	for _, scheme := range signatureSchemes() {
		block, ok := blocks[scheme.id]
		if !ok {
			continue
		}
		c.signing.Schemes = append(c.signing.Schemes, scheme.name)
		certs, err := signerCertificates(block)
		if err != nil {
			return fmt.Errorf("%s signature: %w", scheme.name, err)
		}
		c.addCerts(certs)
	}
	return nil
}
