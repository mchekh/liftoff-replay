package liftoffreplay

import (
	"encoding/binary"
	"io"
)

type RecordEncoder[T ReplayRecordConstraint] interface {
	Format() string
	Encode(out io.Writer, src Iterator[T]) error
}

func ConvertLiftoffReplayFlat(in io.Reader, out io.Writer, enc RecordEncoder[ReplayRecordFlat]) error {
	rr, err := newFlatIterator(in)
	if err != nil {
		return err
	}
	return enc.Encode(out, rr)
}

func ConvertLiftoffReplayNested(in io.Reader, out io.Writer, enc RecordEncoder[ReplayRecordNested]) error {
	rr, err := newFlatIterator(in)
	if err != nil {
		return err
	}
	nested := NewMapIterator[ReplayRecordFlat, ReplayRecordNested](rr, ToNestedRecord)
	return enc.Encode(out, nested)
}

func newFlatIterator(in io.Reader) (Iterator[ReplayRecordFlat], error) {
	decoded, err := ExtractReplayBinaryData(in)
	if err != nil {
		return nil, err
	}
	return NewReplayRecordReader(decoded, binary.LittleEndian), nil
}
