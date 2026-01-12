package telemetry_encoder

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

type TelemetrySource interface {
	NextFrame() ([]byte, error)
}

type Reader[T any] struct {
	src     TelemetrySource
	dec     *StructDecoder[T]
	minSize int

	reuse T
}

func NewReader[T any](src TelemetrySource, schema *TelemetrySchema) (*Reader[T], error) {
	if src == nil {
		return nil, fmt.Errorf("reader: source is nil")
	}
	if schema == nil {
		return nil, fmt.Errorf("reader: schema is nil")
	}

	dec, err := NewStructDecoder[T](schema, binary.LittleEndian, true)
	if err != nil {
		return nil, err
	}

	return &Reader[T]{
		src:     src,
		dec:     dec,
		minSize: schema.Size(),
	}, nil
}

func (r *Reader[T]) NextInto(dst *T) error {
	if dst == nil {
		return fmt.Errorf("reader: dst is nil")
	}

	frame, err := r.src.NextFrame()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return io.EOF
		}
		return err
	}

	if len(frame) < r.minSize {
		return fmt.Errorf("reader: frame too short: got %d, need >= %d", len(frame), r.minSize)
	}

	return r.dec.DecodeInto(frame, dst)
}

func (r *Reader[T]) Next() (T, error) {
	var zero T

	if err := r.NextInto(&r.reuse); err != nil {
		if errors.Is(err, io.EOF) {
			return zero, io.EOF
		}
		return zero, err
	}

	return r.reuse, nil
}
