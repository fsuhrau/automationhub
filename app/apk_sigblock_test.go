package app

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"
)

func le32(v uint32) []byte { return binary.LittleEndian.AppendUint32(nil, v) }
func le64(v uint64) []byte { return binary.LittleEndian.AppendUint64(nil, v) }

// lp prefixes the data with its length
func lp(parts ...[]byte) []byte {
	data := bytes.Join(parts, nil)
	return append(le32(uint32(len(data))), data...)
}

// certSequence is the length prefixed list of certificates
func certSequence(ders ...[]byte) []byte {
	var parts [][]byte
	for _, der := range ders {
		parts = append(parts, lp(der))
	}
	return lp(parts...)
}

func signerValue(ders ...[]byte) []byte {
	signedData := lp(lp([]byte("digests")), certSequence(ders...))
	return lp(lp(signedData))
}

// eocd is an end of central directory record with the central directory at offset
func eocd(offset uint32) []byte {
	rec := make([]byte, eocdSize)
	copy(rec, []byte{0x50, 0x4b, 0x05, 0x06})
	binary.LittleEndian.PutUint32(rec[16:], offset)
	return rec
}

// apkWithBlock is the layout of a signed apk without the zip entries: the
// signing block is right before the central directory
func apkWithBlock(pairs map[uint32][]byte) []byte {
	var body []byte
	for id, value := range pairs {
		body = append(body, le64(uint64(len(value)+4))...)
		body = append(body, le32(id)...)
		body = append(body, value...)
	}
	apk := append(make([]byte, 40), le64(uint64(len(body)+24))...)
	apk = append(apk, body...)
	apk = append(apk, le64(uint64(len(body)+24))...)
	apk = append(apk, sigBlockMagic...)
	cdOffset := len(apk)
	apk = append(apk, []byte("central directory")...)
	return append(apk, eocd(uint32(cdOffset))...)
}

func TestAPKSigningBlocks(t *testing.T) {
	pairs := map[uint32][]byte{apkSigV2: []byte("two"), apkSigV3: []byte("three"), 0x42: nil}
	apk := apkWithBlock(pairs)
	got, err := apkSigningBlocks(bytes.NewReader(apk), int64(len(apk)))
	if err != nil {
		t.Fatal(err)
	}
	if string(got[apkSigV2]) != "two" || string(got[apkSigV3]) != "three" || len(got) != 3 {
		t.Errorf("blocks = %v", got)
	}
}

func TestAPKSigningBlocksWithoutBlock(t *testing.T) {
	tests := map[string][]byte{
		"magic missing":         append(append(make([]byte, 100), []byte("central directory")...), eocd(100)...),
		"central dir at start":  eocd(0),
		"empty block":           apkWithBlock(nil),
		"central dir too small": eocd(31),
	}
	for name, apk := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := apkSigningBlocks(bytes.NewReader(apk), int64(len(apk)))
			if err != nil || len(got) != 0 {
				t.Errorf("blocks = %v, err = %v", got, err)
			}
		})
	}
}

func TestAPKSigningBlocksErrors(t *testing.T) {
	badSize := apkWithBlock(map[uint32][]byte{1: []byte("x")})
	binary.LittleEndian.PutUint64(badSize[len(badSize)-len("central directory")-eocdSize-24:], 1<<40)
	badEntry := apkWithBlock(map[uint32][]byte{1: []byte("xxxxxxxx")})
	binary.LittleEndian.PutUint64(badEntry[40+8:], 1<<20)
	tests := map[string][]byte{
		"no end record":       []byte("not a zip"),
		"truncated end":       append(bytes.Repeat([]byte{0}, 30), 0x50, 0x4b, 0x05, 0x06, 0),
		"bad block size":      badSize,
		"bad entry":           badEntry,
		"central dir missing": eocd(1000),
	}
	for name, apk := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := apkSigningBlocks(bytes.NewReader(apk), int64(len(apk))); err == nil {
				t.Error("no error")
			}
		})
	}
}

// failingReader fails the nth ReadAt call
type failingReader struct {
	r      io.ReaderAt
	calls  int
	failAt int
}

func (f *failingReader) ReadAt(p []byte, off int64) (int, error) {
	f.calls++
	if f.calls == f.failAt {
		return 0, errors.New("disk failure")
	}
	return f.r.ReadAt(p, off)
}

func TestAPKSigningBlocksReadErrors(t *testing.T) {
	apk := apkWithBlock(map[uint32][]byte{1: []byte("x")})
	for failAt := 1; failAt <= 3; failAt++ {
		r := &failingReader{r: bytes.NewReader(apk), failAt: failAt}
		if _, err := apkSigningBlocks(r, int64(len(apk))); err == nil || !strings.Contains(err.Error(), "disk failure") {
			t.Errorf("read %d: err = %v", failAt, err)
		}
	}
}

// zip64APK has the offset of the central directory in the zip64 records
func zip64APK(cdOffset uint64, locator bool) []byte {
	apk := make([]byte, 64)
	rec := make([]byte, 56)
	binary.LittleEndian.PutUint64(rec[48:], cdOffset)
	apk = append(apk, rec...)
	if locator {
		loc := make([]byte, 20)
		binary.LittleEndian.PutUint32(loc, zip64LocatorSig)
		binary.LittleEndian.PutUint64(loc[8:], 64)
		apk = append(apk, loc...)
	}
	return append(apk, eocd(zip64Marker)...)
}

func TestCentralDirectoryOffsetZip64(t *testing.T) {
	apk := zip64APK(12345, true)
	got, err := centralDirectoryOffset(bytes.NewReader(apk), int64(len(apk)))
	if err != nil || got != 12345 {
		t.Errorf("offset = %d, err = %v", got, err)
	}
}

func TestCentralDirectoryOffsetZip64Errors(t *testing.T) {
	noLocator := zip64APK(1, false)
	badPointer := zip64APK(1, true)
	binary.LittleEndian.PutUint64(badPointer[len(badPointer)-eocdSize-12:], 1<<30)
	tooShort := eocd(zip64Marker)
	for name, apk := range map[string][]byte{"no locator": noLocator, "locator points nowhere": badPointer, "no room for locator": tooShort} {
		t.Run(name, func(t *testing.T) {
			if _, err := centralDirectoryOffset(bytes.NewReader(apk), int64(len(apk))); err == nil {
				t.Error("no error")
			}
		})
	}
}

func TestLengthPrefixed(t *testing.T) {
	got, err := lengthPrefixed(append(lp([]byte("ab")), lp([]byte("c"))...))
	if err != nil || len(got) != 2 || string(got[0]) != "ab" || string(got[1]) != "c" {
		t.Errorf("values = %q, err = %v", got, err)
	}
	for name, data := range map[string][]byte{"short prefix": {1, 0}, "value too long": append(le32(9), 1, 2)} {
		if _, err := lengthPrefixed(data); !errors.Is(err, errTruncated) {
			t.Errorf("%s: err = %v", name, err)
		}
	}
}

func TestSignerCertificates(t *testing.T) {
	a, b := testCertificate(t, "A"), testCertificate(t, "B")
	got, err := signerCertificates(signerValue(a.Raw, b.Raw))
	if err != nil || len(got) != 2 || got[1].Subject.CommonName != "B" {
		t.Errorf("certs = %v, err = %v", got, err)
	}
}

func TestSignerCertificatesErrors(t *testing.T) {
	tests := map[string][]byte{
		"empty block":        nil,
		"signers truncated":  lp([]byte{1, 2}),
		"no signed data":     lp(lp()),
		"no digests":         lp(lp(lp())),
		"no certificates":    lp(lp(lp(lp([]byte("d"))))),
		"certificates break": lp(lp(lp(lp([]byte("d")), lp([]byte{1})))),
		"bad certificate":    lp(lp(lp(lp([]byte("d")), certSequence([]byte("not a certificate"))))),
	}
	for name, block := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := signerCertificates(block); err == nil {
				t.Error("no error")
			}
		})
	}
}

func TestSigningPairsKeepsEveryId(t *testing.T) {
	block := append(append(le64(5), le32(7)...), 'x')
	got, err := signingPairs(block)
	if err != nil || !reflect.DeepEqual(got, map[uint32][]byte{7: []byte("x")}) {
		t.Errorf("pairs = %v, err = %v", got, err)
	}
}
