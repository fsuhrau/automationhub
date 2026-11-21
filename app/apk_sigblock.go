package app

import (
	"bytes"
	"crypto/x509"
	"encoding/binary"
	"errors"
	"io"
)

const (
	eocdSize        = 22
	zip64Marker     = 0xffffffff
	zip64LocatorSig = 0x07064b50
	sigBlockMagic   = "APK Sig Block 42"
)

var errTruncated = errors.New("truncated")

// apkSigningBlocks returns the id-value pairs of the APK Signing Block, which
// sits right before the zip central directory
func apkSigningBlocks(r io.ReaderAt, size int64) (map[uint32][]byte, error) {
	cdOffset, err := centralDirectoryOffset(r, size)
	if err != nil {
		return nil, err
	}
	block, err := readSigningBlock(r, cdOffset)
	if err != nil {
		return nil, err
	}
	return signingPairs(block)
}

// centralDirectoryOffset reads the offset from the end of central directory
// record in the tail of the file
func centralDirectoryOffset(r io.ReaderAt, size int64) (int64, error) {
	tail := min(int64(65535+eocdSize), size)
	buf := make([]byte, tail)
	if _, err := r.ReadAt(buf, size-tail); err != nil && !errors.Is(err, io.EOF) {
		return 0, err
	}
	eocd := bytes.LastIndex(buf, []byte{0x50, 0x4b, 0x05, 0x06})
	if eocd < 0 || eocd+eocdSize > len(buf) {
		return 0, errors.New("no end of central directory")
	}
	offset := int64(binary.LittleEndian.Uint32(buf[eocd+16:]))
	if offset == zip64Marker {
		return zip64DirectoryOffset(r, buf[:eocd])
	}
	return offset, nil
}

// zip64DirectoryOffset reads the offset from the zip64 end of central
// directory record, which the locator right before the end record points to
func zip64DirectoryOffset(r io.ReaderAt, beforeEOCD []byte) (int64, error) {
	loc := len(beforeEOCD) - 20
	if loc < 0 || binary.LittleEndian.Uint32(beforeEOCD[loc:]) != zip64LocatorSig {
		return 0, errors.New("no zip64 locator")
	}
	rec := make([]byte, 56)
	if _, err := r.ReadAt(rec, int64(binary.LittleEndian.Uint64(beforeEOCD[loc+8:]))); err != nil {
		return 0, err
	}
	return int64(binary.LittleEndian.Uint64(rec[48:])), nil
}

// readSigningBlock returns the pairs part of the block before the central
// directory, or nothing when the apk has no signing block
func readSigningBlock(r io.ReaderAt, cdOffset int64) ([]byte, error) {
	if cdOffset < 32 {
		return nil, nil
	}
	footer := make([]byte, 24)
	if _, err := r.ReadAt(footer, cdOffset-24); err != nil {
		return nil, err
	}
	if string(footer[8:]) != sigBlockMagic {
		return nil, nil
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
	return block, nil
}

// signingPairs splits the pairs of a signing block: uint64 length, uint32 id, value
func signingPairs(block []byte) (map[uint32][]byte, error) {
	pairs := map[uint32][]byte{}
	for len(block) >= 12 {
		n := binary.LittleEndian.Uint64(block)
		if n < 4 || n > uint64(len(block)-8) {
			return nil, errors.New("bad signing block entry")
		}
		pairs[binary.LittleEndian.Uint32(block[8:])] = block[12 : 8+n]
		block = block[8+n:]
	}
	return pairs, nil
}

// lengthPrefixed splits a sequence of uint32 length prefixed values
func lengthPrefixed(data []byte) ([][]byte, error) {
	var out [][]byte
	for len(data) > 0 {
		value, rest, err := single(data)
		if err != nil {
			return nil, err
		}
		out = append(out, value)
		data = rest
	}
	return out, nil
}

// single takes the first uint32 length prefixed value and returns the rest
func single(data []byte) ([]byte, []byte, error) {
	if len(data) < 4 {
		return nil, nil, errTruncated
	}
	n := binary.LittleEndian.Uint32(data)
	if uint64(n) > uint64(len(data)-4) {
		return nil, nil, errTruncated
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
		signerCerts, err := certificatesOfSigner(signer)
		if err != nil {
			return nil, err
		}
		certs = append(certs, signerCerts...)
	}
	return certs, nil
}

func certificatesOfSigner(signer []byte) ([]*x509.Certificate, error) {
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
	return parseCertificates(ders)
}

func parseCertificates(ders [][]byte) ([]*x509.Certificate, error) {
	var certs []*x509.Certificate
	for _, der := range ders {
		c, err := x509.ParseCertificate(der)
		if err != nil {
			return nil, err
		}
		certs = append(certs, c)
	}
	return certs, nil
}
