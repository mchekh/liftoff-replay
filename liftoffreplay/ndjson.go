package liftoffreplay

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
)

type NdJsonRecordEncoder[T ReplayRecordConstraint] struct{}

func (e *NdJsonRecordEncoder[T]) Format() string { return "ndjson" }

func (e *NdJsonRecordEncoder[T]) Encode(out io.Writer, src Iterator[T]) error {
	bw := bufio.NewWriterSize(out, 256*1024)
	defer bw.Flush()

	enc := json.NewEncoder(bw)
	enc.SetEscapeHTML(false)

	for {
		rec, err := src.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return fmt.Errorf("read record: %w", err)
		}
		if err := enc.Encode(rec); err != nil {
			return fmt.Errorf("encode json: %w", err)
		}
	}
}
