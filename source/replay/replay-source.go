package replay

import (
	"errors"
	"io"
	"os"
)

type ReplayTelemetrySource struct {
	r        io.Reader
	closer   io.Closer
	frameBuf []byte
}

func ReplayTelemetrySourceFromReader(rc io.Reader, frameSize int) (*ReplayTelemetrySource, error) {
	if rc == nil {
		return nil, errors.New("replay source: nil reader")
	}
	if frameSize <= 0 {
		return nil, errors.New("replay source: frameSize must be > 0")
	}

	decoded, err := ExtractReplayBinaryData(rc)
	if err != nil {
		return nil, err
	}

	return &ReplayTelemetrySource{
		r:        decoded,
		frameBuf: make([]byte, frameSize),
	}, nil
}

func ReplayTelemetrySourceFromFile(path string, frameSize int) (*ReplayTelemetrySource, error) {
	if path == "" {
		return nil, errors.New("replay source: empty path")
	}
	if frameSize <= 0 {
		return nil, errors.New("replay source: frameSize must be > 0")
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	decoded, err := ExtractReplayBinaryData(f)
	if err != nil {
		_ = f.Close()
		return nil, err
	}

	return &ReplayTelemetrySource{
		r:        decoded,
		closer:   f,
		frameBuf: make([]byte, frameSize),
	}, nil
}

func (s *ReplayTelemetrySource) NextFrame() ([]byte, error) {
	if s == nil || s.r == nil {
		return nil, errors.New("replay source: not initialized")
	}

	_, err := io.ReadFull(s.r, s.frameBuf)
	if err != nil {
		return nil, err
	}

	return s.frameBuf, nil
}

func (s *ReplayTelemetrySource) Close() error {
	if s == nil || s.closer == nil {
		return nil
	}
	return s.closer.Close()
}
