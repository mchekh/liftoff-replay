package telemetry

import "io"

type FileTelemetrySource struct {
	rc   io.ReadCloser
	buff []byte
}

func NewFileTelemetrySource(rc io.ReadCloser, frameSize int) *FileTelemetrySource {
	return &FileTelemetrySource{
		rc:   rc,
		buff: make([]byte, frameSize),
	}
}

func (s *FileTelemetrySource) NextFrame() ([]byte, error) {
	_, err := s.rc.Read(s.buff)
	if err != nil {
		return nil, err
	}
	return s.buff[:], nil
}

func (s *FileTelemetrySource) Close() error {
	return s.rc.Close()
}
