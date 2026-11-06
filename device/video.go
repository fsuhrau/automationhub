package device

import "context"

// VideoOptions tune a live video stream. Zero values use the defaults of the
// implementation.
type VideoOptions struct {
	// MaxSize limits the longer side of the video in pixels (0 = native size).
	MaxSize int
	// MaxFPS limits the frame rate (0 = unlimited).
	MaxFPS int
	// BitRate in bits per second.
	BitRate int
}

// Codecs of VideoMeta.Codec.
const (
	// CodecH264 packets are H.264 in Annex B format.
	CodecH264 = "h264"
	// CodecMJPEG packets are complete JPEG images, each one a key frame.
	CodecMJPEG = "mjpeg"
)

// VideoMeta describes a video stream.
type VideoMeta struct {
	// Codec is the codec of the packets, CodecH264 or CodecMJPEG.
	Codec string `json:"codec"`
	// Width and Height of the video at the start of the stream. Rotations
	// change them; decoders take the size from the config packets.
	Width  int `json:"width"`
	Height int `json:"height"`
	// ScreenWidth and ScreenHeight are the device resolution in the natural
	// orientation, the coordinate space of device.InputCapable. Clients map a
	// position on the video to device pixels with it, swapping the two when
	// the video orientation differs. Zero when unknown.
	ScreenWidth  int `json:"screenWidth"`
	ScreenHeight int `json:"screenHeight"`
	// ViewOnly streams take no input (they don't implement VideoInput);
	// clients use the control API (device.InputCapable) instead.
	ViewOnly bool `json:"viewOnly,omitempty"`
}

// VideoPacket is one encoded unit of a video stream. For H.264 the data is in
// Annex B format.
type VideoPacket struct {
	// Config packets carry the codec configuration (SPS/PPS) and no picture.
	Config   bool
	KeyFrame bool
	// PTS is the presentation time in microseconds (0 for config packets).
	PTS  int64
	Data []byte
}

// VideoStream is a running video stream. ReadPacket blocks until the next
// packet; it returns an error once the stream ended or was closed.
type VideoStream interface {
	Meta() VideoMeta
	ReadPacket() (VideoPacket, error)
	Close() error
}

// VideoStreamRequest asks a node to stream a device's video back to the
// master under StreamID.
type VideoStreamRequest struct {
	DeviceID string
	StreamID string
	Options  VideoOptions
}

// VideoStreamer is implemented by devices that can stream their screen as
// encoded video. The stream ends when ctx is done or it is closed.
type VideoStreamer interface {
	StartVideo(ctx context.Context, opts VideoOptions) (VideoStream, error)
}

// Input event kinds and actions of InputEvent.
const (
	InputTouch      = "touch"
	InputKey        = "key"
	InputText       = "text"
	InputScroll     = "scroll"
	InputResetVideo = "reset_video"

	InputDown   = "down"
	InputMove   = "move"
	InputUp     = "up"
	InputCancel = "cancel"
	// InputPress is a key down followed by a key up.
	InputPress = "press"
)

// InputEvent is low latency input sent over a running video stream.
// Positions are pixels of the current video frame, whose size is Width x
// Height; events for another size are dropped (e.g. during a rotation).
type InputEvent struct {
	Kind      string  `json:"kind"`
	Action    string  `json:"action,omitempty"`
	PointerID int64   `json:"pointerId,omitempty"`
	X         int     `json:"x,omitempty"`
	Y         int     `json:"y,omitempty"`
	Width     int     `json:"width,omitempty"`
	Height    int     `json:"height,omitempty"`
	Key       string  `json:"key,omitempty"`
	Text      string  `json:"text,omitempty"`
	ScrollX   float64 `json:"scrollX,omitempty"`
	ScrollY   float64 `json:"scrollY,omitempty"`
}

// VideoInput is implemented by video streams that inject input through the
// stream's own channel, which is much faster than device.InputCapable.
type VideoInput interface {
	InjectInput(ev InputEvent) error
}
