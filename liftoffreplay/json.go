package liftoffreplay

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
)

type JsonRecordEncoder[T ReplayRecordConstraint] struct{}

func (e *JsonRecordEncoder[T]) Format() string { return "json" }

func (e *JsonRecordEncoder[T]) Encode(out io.Writer, src Iterator[T]) error {

	bw := bufio.NewWriterSize(out, 256*1024)
	defer bw.Flush()

	if _, err := bw.WriteString("["); err != nil {
		return err
	}

	first := true
	for {
		rec, err := src.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read record: %w", err)
		}

		if !first {
			if err := bw.WriteByte(','); err != nil {
				return err
			}
		} else {
			first = false
		}

		b, err := json.Marshal(rec)
		if err != nil {
			return fmt.Errorf("marshal json: %w", err)
		}
		if _, err := bw.Write(b); err != nil {
			return err
		}
	}

	if _, err := bw.WriteString("]\n"); err != nil {
		return err
	}
	return nil
}
