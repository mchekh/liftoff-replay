package liftoffreplay

import (
	"bufio"
	"encoding/csv"
	"io"
	"strconv"
)

type CsvRecordEncoder struct{}

func (e *CsvRecordEncoder) Format() string { return "csv" }

func (e *CsvRecordEncoder) Encode(out io.Writer, src Iterator[ReplayRecordFlat]) error {
	bw := bufio.NewWriterSize(out, 256*1024)
	defer bw.Flush()

	w := csv.NewWriter(bw)
	defer w.Flush()

	if err := w.Write([]string{
		"pos_x", "pos_y", "pos_z",
		"rot_x", "rot_y", "rot_z", "rot_w",
		"throttle", "yaw", "pitch", "roll",
		"time_sec",
	}); err != nil {
		return err
	}

	for {
		rec, err := src.Next()
		if err == io.EOF {
			return w.Error()
		}
		if err != nil {
			return err
		}

		row := []string{
			f32(rec.PosX),
			f32(rec.PosY),
			f32(rec.PosZ),
			f32(rec.RotX),
			f32(rec.RotY),
			f32(rec.RotZ),
			f32(rec.RotW),
			f32(rec.Throttle),
			f32(rec.Yaw),
			f32(rec.Pitch),
			f32(rec.Roll),
			f32(rec.TimeSec),
		}

		if err := w.Write(row); err != nil {
			return err
		}
	}
}

func f32(v float32) string {
	return strconv.FormatFloat(float64(v), 'f', -1, 32)
}
