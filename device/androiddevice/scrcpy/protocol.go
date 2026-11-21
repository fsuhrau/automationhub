package scrcpy

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/fsuhrau/automationhub/device"
)

const (
	// maxPacketSize guards against a corrupt stream; real packets are far smaller
	maxPacketSize = 16 << 20

	codecH264 = 0x68323634 // "h264"
	codecH265 = 0x68323635 // "h265"
	codecAV1  = 0x00617631 // "av1"

	flagConfig   = uint64(1) << 63
	flagKeyFrame = uint64(1) << 62
)

// readHeader reads the codec meta: codec id, width and height (big endian).
func readHeader(r io.Reader) (string, int, int, error) {
	var header [12]byte
	if _, err := io.ReadFull(r, header[:4]); err != nil {
		return "", 0, 0, fmt.Errorf("read scrcpy codec: %w", err)
	}
	codec, err := codecName(binary.BigEndian.Uint32(header[:4]))
	if err != nil {
		return "", 0, 0, err
	}
	if _, err := io.ReadFull(r, header[4:]); err != nil {
		return "", 0, 0, fmt.Errorf("read scrcpy video size: %w", err)
	}
	return codec, int(binary.BigEndian.Uint32(header[4:8])), int(binary.BigEndian.Uint32(header[8:12])), nil
}

func codecName(id uint32) (string, error) {
	switch id {
	case 0:
		return "", errors.New("scrcpy disabled the video stream")
	case 1:
		return "", errors.New("scrcpy video configuration error, see the device log")
	case codecH264:
		return "h264", nil
	case codecH265:
		return "h265", nil
	case codecAV1:
		return "av1", nil
	}
	return "", fmt.Errorf("unknown scrcpy codec id %#x", id)
}

// readPacket reads one frame header (pts and flags, size) and its payload.
func readPacket(r io.Reader) (device.VideoPacket, error) {
	var header [12]byte
	if _, err := io.ReadFull(r, header[:]); err != nil {
		return device.VideoPacket{}, err
	}
	ptsFlags := binary.BigEndian.Uint64(header[:8])
	size := binary.BigEndian.Uint32(header[8:])
	if size == 0 || size > maxPacketSize {
		return device.VideoPacket{}, fmt.Errorf("invalid scrcpy packet size %d", size)
	}
	data := make([]byte, size)
	if _, err := io.ReadFull(r, data); err != nil {
		return device.VideoPacket{}, err
	}
	p := device.VideoPacket{Data: data}
	if ptsFlags&flagConfig != 0 {
		p.Config = true
	} else {
		p.KeyFrame = ptsFlags&flagKeyFrame != 0
		p.PTS = int64(ptsFlags &^ (flagConfig | flagKeyFrame))
	}
	return p, nil
}
