package app

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	"github.com/fullsailor/pkcs7"
)

func TestIsV1SignatureFile(t *testing.T) {
	for name, want := range map[string]bool{
		"META-INF/CERT.RSA": true, "META-INF/cert.dsa": true, "META-INF/X.EC": true,
		"META-INF/MANIFEST.MF": false, "META-INF/sub/CERT.RSA": false, "other/CERT.RSA": false, "META-INF/CERT": false,
	} {
		if got := isV1SignatureFile(name); got != want {
			t.Errorf("isV1SignatureFile(%s) = %v, want %v", name, got, want)
		}
	}
}

func v1Signature(t *testing.T, commonName string) []byte {
	t.Helper()
	data, err := pkcs7.DegenerateCertificate(testCertificate(t, commonName).Raw)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestReadAPKSigningAllSchemes(t *testing.T) {
	debug, release := testCertificate(t, "Android Debug"), testCertificate(t, "Release")
	apk := apkWithBlock(map[uint32][]byte{
		apkSigV2:  signerValue(debug.Raw, release.Raw),
		apkSigV3:  signerValue(debug.Raw),
		apkSigV31: signerValue(release.Raw),
	})
	_, files := openZip(t, map[string][]byte{
		"META-INF/CERT.RSA":     v1Signature(t, "Legacy"),
		"META-INF/OTHER.RSA":    []byte("not pkcs7"),
		"META-INF/MANIFEST.MF":  []byte("manifest"),
		"META-INF/nested/X.RSA": []byte("ignored"),
		"classes.dex":           []byte("dex"),
	}, "META-INF/BROKEN.EC")
	signing, err := readAPKSigning(bytes.NewReader(apk), int64(len(apk)), files)
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"v1", "v2", "v3", "v3.1"}; !reflect.DeepEqual(signing.Schemes, want) {
		t.Errorf("schemes = %v, want %v", signing.Schemes, want)
	}
	if len(signing.Certificates) != 3 {
		t.Fatalf("certificates = %d, want 3 without duplicates", len(signing.Certificates))
	}
	if !signing.Certificates[1].Debug || signing.Certificates[2].Debug {
		t.Errorf("certificates = %+v", signing.Certificates)
	}
}

func TestReadAPKSigningWithoutSignature(t *testing.T) {
	apk := apkWithBlock(nil)
	signing, err := readAPKSigning(bytes.NewReader(apk), int64(len(apk)), nil)
	if err != nil || len(signing.Schemes) != 0 || len(signing.Certificates) != 0 {
		t.Errorf("signing = %+v, err = %v", signing, err)
	}
}

func TestReadAPKSigningKeepsV1WhenBlockIsUnreadable(t *testing.T) {
	_, files := openZip(t, map[string][]byte{"META-INF/CERT.RSA": v1Signature(t, "Legacy")})
	signing, err := readAPKSigning(bytes.NewReader([]byte("no zip")), 6, files)
	if err == nil {
		t.Error("no error")
	}
	if !reflect.DeepEqual(signing.Schemes, []string{"v1"}) || len(signing.Certificates) != 1 {
		t.Errorf("signing = %+v", signing)
	}
}

func TestReadAPKSigningNamesTheSchemeThatFails(t *testing.T) {
	apk := apkWithBlock(map[uint32][]byte{apkSigV3: []byte("broken")})
	signing, err := readAPKSigning(bytes.NewReader(apk), int64(len(apk)), nil)
	if err == nil || !strings.HasPrefix(err.Error(), "v3 signature: ") {
		t.Errorf("err = %v", err)
	}
	if !reflect.DeepEqual(signing.Schemes, []string{"v3"}) {
		t.Errorf("schemes = %v", signing.Schemes)
	}
}
