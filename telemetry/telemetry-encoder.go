package telemetry

import (
	"encoding/binary"
	"fmt"
	"math"
	"reflect"
	"strings"
)

type TelemetryEncoder[T any] struct {
	order binary.ByteOrder
	binds []bind

	recordSize int
}

type bind struct {
	key       string
	offset    int
	prim      Primitive
	fieldIdx  int
	isPointer bool
	baseType  reflect.Type
	size      int
}

func NewTelemetryEncoder[T any](schema *TelemetrySchema, order binary.ByteOrder, allowOptional bool) (*TelemetryEncoder[T], error) {
	if schema == nil {
		return nil, fmt.Errorf("decoder: schema is nil")
	}
	if order == nil {
		order = binary.LittleEndian
	}

	var zero T
	rt := reflect.TypeOf(zero)
	if rt.Kind() == reflect.Pointer {
		rt = rt.Elem()
	}
	if rt.Kind() != reflect.Struct {
		return nil, fmt.Errorf("decoder: T must be a struct, got %s", rt.Kind())
	}

	binds := make([]bind, 0, rt.NumField())

	for i := 0; i < rt.NumField(); i++ {
		sf := rt.Field(i)
		tag := strings.TrimSpace(sf.Tag.Get("telemetry"))
		if tag == "" {
			continue
		}

		fi, ok := schema.Field(tag)
		if !ok {
			if allowOptional {
				// Absent in runtime schema:
				// - value fields => remain zero
				// - pointer fields => remain nil
				continue
			}
			return nil, fmt.Errorf("decoder: field tag is not in the schema: %s", tag)
		}

		ft := sf.Type
		isPtr := ft.Kind() == reflect.Pointer
		base := ft
		if isPtr {
			base = ft.Elem()
		}

		if err := ensureCompatible(base, fi.Desc.Type); err != nil {
			return nil, fmt.Errorf("decoder: struct field %s tag %q: %w", sf.Name, tag, err)
		}

		sz, _ := fi.Desc.Size()

		binds = append(binds, bind{
			key:       tag,
			offset:    fi.Offset,
			prim:      fi.Desc.Type,
			fieldIdx:  i,
			isPointer: isPtr,
			baseType:  base,
			size:      sz,
		})
	}

	return &TelemetryEncoder[T]{order: order, binds: binds, recordSize: schema.Size()}, nil
}

func (d *TelemetryEncoder[T]) RecordSize() int {
	return d.recordSize
}

func (d *TelemetryEncoder[T]) DecodeInto(buf []byte, out *T) error {
	if out == nil {
		return fmt.Errorf("decoder: out is nil")
	}

	v := reflect.ValueOf(out).Elem()

	for _, b := range d.binds {
		end := b.offset + b.size
		if b.offset < 0 || end > len(buf) {
			return fmt.Errorf("decoder: buffer too short for %q (need [%d:%d], have %d)",
				b.key, b.offset, end, len(buf))
		}

		fv := v.Field(b.fieldIdx)

		if b.isPointer {
			if fv.IsNil() {
				fv.Set(reflect.New(b.baseType))
			}
			fv = fv.Elem()
		}

		switch b.prim {
		case PrimByte:
			fv.SetUint(uint64(buf[b.offset]))

		case PrimFloat32:
			u := d.order.Uint32(buf[b.offset : b.offset+4])
			fv.SetFloat(float64(math.Float32frombits(u)))

		default:
			return fmt.Errorf("decoder: unsupported primitive %v for %q", b.prim, b.key)
		}
	}
	return nil
}

func (d *TelemetryEncoder[T]) Decode(buf []byte) (T, error) {
	var out T
	if err := d.DecodeInto(buf, &out); err != nil {
		return out, err
	}
	return out, nil
}

func ensureCompatible(goType reflect.Type, prim Primitive) error {
	switch prim {
	case PrimByte:
		if goType.Kind() != reflect.Uint8 {
			return fmt.Errorf("prim byte requires uint8, got %s", goType)
		}
	case PrimFloat32:
		if goType.Kind() != reflect.Float32 {
			return fmt.Errorf("prim float32 requires float32, got %s", goType)
		}
	default:
		return fmt.Errorf("unknown primitive %v", prim)
	}
	return nil
}
