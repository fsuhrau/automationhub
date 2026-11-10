package app

import (
	"archive/zip"
	"bufio"
	"bytes"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/x509"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/fullsailor/pkcs7"
)

var (
	badgingAttrRegex  = regexp.MustCompile(`([A-Za-z][\w-]*)='((?:[^'\\]|\\.)*)'`)
	badgingQuoteRegex = regexp.MustCompile(`'((?:[^'\\]|\\.)*)'`)
	iconDensityRegex  = regexp.MustCompile(`^application-icon-(\d+)$`)
)

// parseBadging reads the output of "aapt dump badging"
func parseBadging(output []byte) (*AndroidInfo, map[int]string) {
	info := &AndroidInfo{}
	icons := map[int]string{}
	features := map[string]bool{}
	var featureOrder []string
	addFeature := func(name string, required bool) {
		if name == "" {
			return
		}
		if _, ok := features[name]; !ok {
			featureOrder = append(featureOrder, name)
		}
		features[name] = features[name] || required
	}
	scanner := bufio.NewScanner(bytes.NewReader(output))
	scanner.Buffer(make([]byte, 1024*1024), 16*1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		key, rest, found := strings.Cut(line, ":")
		if !found {
			switch {
			case line == "application-debuggable":
				info.Debuggable = true
			case line == "leanback-launcher":
				info.Leanback = true
			case strings.HasPrefix(line, "testOnly="):
				info.TestOnly = true
			}
			continue
		}
		rest = strings.TrimSpace(rest)
		attrs := map[string]string{}
		for _, m := range badgingAttrRegex.FindAllStringSubmatch(rest, -1) {
			attrs[m[1]] = m[2]
		}
		value := func() string {
			if m := badgingQuoteRegex.FindStringSubmatch(rest); m != nil {
				return m[1]
			}
			return rest
		}
		list := func() []string {
			var out []string
			for _, m := range badgingQuoteRegex.FindAllStringSubmatch(rest, -1) {
				out = append(out, m[1])
			}
			return out
		}
		switch key {
		case "package":
			info.PackageName = attrs["name"]
			info.VersionCode = attrs["versionCode"]
			info.VersionName = attrs["versionName"]
			info.CompileSdk = attrs["compileSdkVersion"]
			info.CompileSdkCodename = attrs["compileSdkVersionCodename"]
			info.PlatformBuildVersion = attrs["platformBuildVersionName"]
		case "sdkVersion", "minSdkVersion":
			info.MinSdk = value()
		case "targetSdkVersion":
			info.TargetSdk = value()
		case "maxSdkVersion":
			info.MaxSdk = value()
		case "application-label":
			info.Label = value()
		case "application":
			if info.Label == "" {
				info.Label = attrs["label"]
			}
			if icon := attrs["icon"]; icon != "" {
				if _, ok := icons[0]; !ok {
					icons[0] = icon
				}
			}
		case "launchable-activity":
			if info.LaunchActivity == "" {
				info.LaunchActivity = attrs["name"]
			}
		case "uses-permission", "uses-permission-sdk-23", "uses-permission-sdk-m":
			if name := attrs["name"]; name != "" {
				info.Permissions = append(info.Permissions, AndroidPermission{Name: name, MaxSdk: attrs["maxSdkVersion"]})
			}
		case "uses-feature":
			addFeature(attrs["name"], attrs["required"] != "false")
		case "uses-feature-not-required":
			addFeature(attrs["name"], false)
		case "uses-gl-es":
			info.GLES = glesVersion(value())
		case "native-code", "alt-native-code":
			info.ABIs = append(info.ABIs, list()...)
		case "locales":
			for _, l := range list() {
				if l != "--_--" {
					info.Locales = append(info.Locales, l)
				}
			}
		case "densities":
			info.Densities = list()
		case "supports-screens":
			info.Screens = list()
		default:
			if m := iconDensityRegex.FindStringSubmatch(key); m != nil {
				d, _ := strconv.Atoi(m[1])
				icons[d] = value()
			}
		}
	}
	for _, name := range featureOrder {
		info.Features = append(info.Features, AndroidFeature{Name: name, Required: features[name]})
	}
	info.Permissions = uniquePermissions(info.Permissions)
	return info, icons
}

func uniquePermissions(in []AndroidPermission) []AndroidPermission {
	seen := map[string]bool{}
	var out []AndroidPermission
	for _, p := range in {
		if !seen[p.Name] {
			seen[p.Name] = true
			out = append(out, p)
		}
	}
	return out
}

// glesVersion turns aapt's '0x00030000' into 3.0
func glesVersion(hex string) string {
	v, err := strconv.ParseUint(strings.TrimPrefix(hex, "0x"), 16, 32)
	if err != nil {
		return hex
	}
	return fmt.Sprintf("%d.%d", v>>16, v&0xffff)
}

// apkIcon picks the biggest png icon. Adaptive icons are xml, so a png with
// the same name in a density folder is taken instead.
func apkIcon(files []*zip.File, icons map[int]string) string {
	densities := make([]int, 0, len(icons))
	for d := range icons {
		densities = append(densities, d)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(densities)))
	var names []string
	for _, d := range densities {
		path := icons[d]
		if strings.HasSuffix(path, ".png") || strings.HasSuffix(path, ".webp") {
			if findFile(files, path) != nil {
				return path
			}
		}
		base := path[strings.LastIndex(path, "/")+1:]
		names = append(names, strings.TrimSuffix(base, ".xml"))
	}
	best, bestRank := "", -1
	for _, f := range files {
		if !strings.HasPrefix(f.Name, "res/mipmap-") && !strings.HasPrefix(f.Name, "res/drawable-") {
			continue
		}
		base := f.Name[strings.LastIndex(f.Name, "/")+1:]
		ext := ""
		if i := strings.LastIndex(base, "."); i > 0 {
			base, ext = base[:i], base[i:]
		}
		if ext != ".png" && ext != ".webp" {
			continue
		}
		for _, n := range names {
			if base == n {
				if rank := densityRank(f.Name); rank > bestRank {
					best, bestRank = f.Name, rank
				}
			}
		}
	}
	return best
}

func densityRank(path string) int {
	for i, d := range []string{"xxxhdpi", "xxhdpi", "xhdpi", "hdpi", "mdpi", "ldpi"} {
		if strings.Contains(path, "-"+d) {
			return 10 - i
		}
	}
	return 0
}

const (
	apkSigV2  = 0x7109871a
	apkSigV3  = 0xf05368c0
	apkSigV31 = 0x1b93ad61
)

// readAPKSigning finds the signature schemes and signer certificates of an
// apk: the APK Signing Block for v2 and later, META-INF for v1
func readAPKSigning(r io.ReaderAt, size int64, files []*zip.File) (*Signing, error) {
	signing := &Signing{}
	seen := map[string]bool{}
	addCert := func(c *x509.Certificate) {
		cert := certificateOf(c)
		if !seen[cert.SHA256] {
			seen[cert.SHA256] = true
			signing.Certificates = append(signing.Certificates, cert)
		}
	}
	for _, f := range files {
		if !strings.HasPrefix(f.Name, "META-INF/") || strings.Count(f.Name, "/") != 1 {
			continue
		}
		ext := strings.ToUpper(f.Name[strings.LastIndex(f.Name, ".")+1:])
		if ext != "RSA" && ext != "DSA" && ext != "EC" {
			continue
		}
		data, err := readZipFile(f, 1<<20)
		if err != nil {
			continue
		}
		p7, err := pkcs7.Parse(data)
		if err != nil {
			continue
		}
		if len(signing.Schemes) == 0 {
			signing.Schemes = append(signing.Schemes, "v1")
		}
		for _, c := range p7.Certificates {
			addCert(c)
		}
	}
	blocks, err := apkSigningBlocks(r, size)
	if err != nil {
		return signing, err
	}
	for _, scheme := range []struct {
		id   uint32
		name string
	}{{apkSigV2, "v2"}, {apkSigV3, "v3"}, {apkSigV31, "v3.1"}} {
		block, ok := blocks[scheme.id]
		if !ok {
			continue
		}
		signing.Schemes = append(signing.Schemes, scheme.name)
		certs, err := signerCertificates(block)
		if err != nil {
			return signing, fmt.Errorf("%s signature: %w", scheme.name, err)
		}
		for _, c := range certs {
			addCert(c)
		}
	}
	return signing, nil
}

// apkSigningBlocks returns the id-value pairs of the APK Signing Block, which
// sits right before the zip central directory
func apkSigningBlocks(r io.ReaderAt, size int64) (map[uint32][]byte, error) {
	tail := int64(65535 + 22)
	if tail > size {
		tail = size
	}
	buf := make([]byte, tail)
	if _, err := r.ReadAt(buf, size-tail); err != nil && !errors.Is(err, io.EOF) {
		return nil, err
	}
	eocd := bytes.LastIndex(buf, []byte{0x50, 0x4b, 0x05, 0x06})
	if eocd < 0 || eocd+22 > len(buf) {
		return nil, errors.New("no end of central directory")
	}
	cdOffset := int64(binary.LittleEndian.Uint32(buf[eocd+16:]))
	if cdOffset == 0xffffffff {
		// zip64: the offset is in the zip64 end of central directory record
		loc := eocd - 20
		if loc < 0 || binary.LittleEndian.Uint32(buf[loc:]) != 0x07064b50 {
			return nil, errors.New("no zip64 locator")
		}
		rec := make([]byte, 56)
		if _, err := r.ReadAt(rec, int64(binary.LittleEndian.Uint64(buf[loc+8:]))); err != nil {
			return nil, err
		}
		cdOffset = int64(binary.LittleEndian.Uint64(rec[48:]))
	}
	if cdOffset < 32 {
		return map[uint32][]byte{}, nil
	}
	footer := make([]byte, 24)
	if _, err := r.ReadAt(footer, cdOffset-24); err != nil {
		return nil, err
	}
	if string(footer[8:]) != "APK Sig Block 42" {
		return map[uint32][]byte{}, nil
	}
	blockSize := int64(binary.LittleEndian.Uint64(footer))
	if blockSize < 24 || blockSize > cdOffset-8 || blockSize > 64<<20 {
		return nil, errors.New("bad signing block size")
	}
	// the block is: size, pairs, size, magic; blockSize counts all but the first size
	block := make([]byte, blockSize-24)
	if _, err := r.ReadAt(block, cdOffset-blockSize); err != nil {
		return nil, err
	}
	pairs := map[uint32][]byte{}
	for len(block) >= 12 {
		n := binary.LittleEndian.Uint64(block)
		if n < 4 || n > uint64(len(block)-8) {
			return nil, errors.New("bad signing block entry")
		}
		id := binary.LittleEndian.Uint32(block[8:])
		pairs[id] = block[12 : 8+n]
		block = block[8+n:]
	}
	return pairs, nil
}

// lengthPrefixed splits a sequence of uint32 length prefixed values
func lengthPrefixed(data []byte) ([][]byte, error) {
	var out [][]byte
	for len(data) > 0 {
		if len(data) < 4 {
			return nil, errors.New("truncated")
		}
		n := binary.LittleEndian.Uint32(data)
		if uint64(n) > uint64(len(data)-4) {
			return nil, errors.New("truncated")
		}
		out = append(out, data[4:4+n])
		data = data[4+n:]
	}
	return out, nil
}

func single(data []byte) ([]byte, []byte, error) {
	if len(data) < 4 {
		return nil, nil, errors.New("truncated")
	}
	n := binary.LittleEndian.Uint32(data)
	if uint64(n) > uint64(len(data)-4) {
		return nil, nil, errors.New("truncated")
	}
	return data[4 : 4+n], data[4+n:], nil
}

// signerCertificates reads the certificates of all signers of a v2 or v3
// block: signers > signer > signed data > (digests, certificates, ...)
func signerCertificates(block []byte) ([]*x509.Certificate, error) {
	signersSeq, _, err := single(block)
	if err != nil {
		return nil, err
	}
	signers, err := lengthPrefixed(signersSeq)
	if err != nil {
		return nil, err
	}
	var certs []*x509.Certificate
	for _, signer := range signers {
		signedData, _, err := single(signer)
		if err != nil {
			return nil, err
		}
		_, rest, err := single(signedData) // digests
		if err != nil {
			return nil, err
		}
		certSeq, _, err := single(rest)
		if err != nil {
			return nil, err
		}
		ders, err := lengthPrefixed(certSeq)
		if err != nil {
			return nil, err
		}
		for _, der := range ders {
			c, err := x509.ParseCertificate(der)
			if err != nil {
				return nil, err
			}
			certs = append(certs, c)
		}
	}
	return certs, nil
}

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
