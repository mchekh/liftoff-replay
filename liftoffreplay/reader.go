package liftoffreplay

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

type ReplayRecordReader struct {
	br    io.Reader
	order binary.ByteOrder
}

func NewReplayRecordReader(br io.Reader, order binary.ByteOrder) *ReplayRecordReader {
	return &ReplayRecordReader{br: br, order: order}
}

func (rr *ReplayRecordReader) Next() (ReplayRecordFlat, error) {
	var rec ReplayRecordFlat
	if err := binary.Read(rr.br, rr.order, &rec); err != nil {
		if err == io.EOF {
			return ReplayRecordFlat{}, io.EOF
		}
		if errors.Is(err, io.ErrUnexpectedEOF) {
			return ReplayRecordFlat{}, fmt.Errorf("truncated record: %w", err)
		}
		return ReplayRecordFlat{}, err
	}
	return rec, nil
}
