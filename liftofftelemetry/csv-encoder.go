package telemetry

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
)

type CSVEncoder struct {
	schema *TelemetrySchema
	w      *csv.Writer
	fields []FieldInstance
}

func NewCSVEncoder(schema *TelemetrySchema, out io.Writer) *CSVEncoder {
	e := &CSVEncoder{
		schema: schema,
		w:      csv.NewWriter(out),
		fields: schema.Fields(),
	}
	e.writeHeader()
	return e
}

func (e *CSVEncoder) writeHeader() {
	header := make([]string, 0, len(e.fields))
	for i := range e.fields {
		fi := e.fields[i]
		if fi.Desc.Type != PrimFloat32 {
			continue
		}
		key := fi.Desc.Alias
		if key == "" {
			key = fi.Desc.Name
		}
		header = append(header, key)
	}
	_ = e.w.Write(header)
}

func (e *CSVEncoder) EncodeAll(src TelemetrySource) error {
	defer e.w.Flush()

	for {
		frame, err := src.NextFrame()
		if err != nil {
			if err == io.EOF {
				return e.w.Error()
			}
			return err
		}

		row := make([]string, 0, len(e.fields))
		for i := range e.fields {
			fi := e.fields[i]
			if fi.Desc.Type != PrimFloat32 {
				continue
			}

			v, err := e.schema.GetF32(fi, frame)
			if err != nil {
				return fmt.Errorf("decode %s: %w", fi.Desc.Name, err)
			}
			row = append(row, strconv.FormatFloat(float64(v), 'f', -1, 32))
		}

		if err := e.w.Write(row); err != nil {
			return err
		}
	}
}
