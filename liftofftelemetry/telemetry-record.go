package telemetry

import (
	"encoding/binary"
	"fmt"
	"math"
)

type TelemetryRecord struct {
	schema *TelemetrySchema
	data   []byte
}

func NewRecord(schema *TelemetrySchema, data []byte) *TelemetryRecord {
	return &TelemetryRecord{
		schema: schema,
		data:   data,
	}
}

func (r *TelemetryRecord) GetF32(name string) (float32, bool, error) {
	fi, ok := r.schema.Field(name)
	if !ok {
		return 0, false, nil
	}

	if fi.Desc.Type != PrimFloat32 {
		return 0, true, fmt.Errorf("field %q is %v, expected float32", name, fi.Desc.Type)
	}

	off := fi.Offset
	if off+4 > len(r.data) {
		return 0, true, fmt.Errorf("buffer too short for field %q (need 4 bytes at offset %d)", name, off)
	}

	u := binary.LittleEndian.Uint32(r.data[off : off+4])
	return math.Float32frombits(u), true, nil
}

func (r *TelemetryRecord) getByte(name string) (byte, bool, error) {
	fi, ok := r.schema.Field(name)
	if !ok {
		return 0, false, nil
	}

	if fi.Desc.Type != PrimByte {
		return 0, true, fmt.Errorf("field %q is %v, expected byte", name, fi.Desc.Type)
	}

	off := fi.Offset
	if off+1 > len(r.data) {
		return 0, true, fmt.Errorf("buffer too short for field %q (need 1 byte at offset %d)", name, off)
	}

	return r.data[off], true, nil
}

func (r *TelemetryRecord) assertQuadMotors() error {
	cnt, ok, err := r.getByte("MotorRPMCount")
	if err != nil {
		return err
	}
	if !ok {
		return nil // schema does not include motors
	}
	if cnt != 4 {
		return fmt.Errorf("expected MotorRPMCount=4, got %d", cnt)
	}
	return nil
}
