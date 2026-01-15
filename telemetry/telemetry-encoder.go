package telemetry

import (
	"encoding/binary"
	"fmt"
	"math"
	"reflect"
	"strings"
)

type TelemetryEncoder[T any] struct {
	order         binary.ByteOrder
	binds         []bind
	recordSize    int
	allowOptional bool
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
		return nil, fmt.Errorf("telemetry encoder: schema is nil")
	}
	if order == nil {
		order = binary.LittleEndian
	}

	var zero T
	rt := reflect.TypeOf(zero)
	if rt == nil {
		return nil, fmt.Errorf("telemetry encoder: T has nil type (unexpected)")
	}
	if rt.Kind() == reflect.Pointer {
		rt = rt.Elem()
	}
	if rt.Kind() != reflect.Struct {
		return nil, fmt.Errorf("telemetry encoder: T must be a struct, got %s", rt.Kind())
	}

	binds := make([]bind, 0, rt.NumField())
	boundKeys := make(map[string]struct{}, rt.NumField())

	for i := 0; i < rt.NumField(); i++ {
		sf := rt.Field(i)
		tag := strings.TrimSpace(sf.Tag.Get("telemetry"))
		if tag == "" {
			continue
		}

		fi, ok := schema.Field(tag)
		if !ok {
			if allowOptional {
				// Tag exists in struct, but not in runtime schema -> ignore.
				continue
			}
			return nil, fmt.Errorf("telemetry encoder: struct tag not in schema: %s", tag)
		}

		ft := sf.Type
		isPtr := ft.Kind() == reflect.Pointer
		base := ft
		if isPtr {
			base = ft.Elem()
		}

		if err := ensureCompatible(base, fi.Desc.Type); err != nil {
			return nil, fmt.Errorf("telemetry encoder: struct field %s tag %q: %w", sf.Name, tag, err)
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
		boundKeys[tag] = struct{}{}
	}

	// Strict mode: struct must cover the full schema.
	if !allowOptional {
		for _, field := range schema.Fields() {
			if _, ok := boundKeys[field.Desc.Name]; !ok {
				return nil, fmt.Errorf("telemetry encoder: schema field %q is missing in struct T", field.Desc.Name)
			}
		}
	}

	return &TelemetryEncoder[T]{
		order:         order,
		binds:         binds,
		recordSize:    schema.Size(),
		allowOptional: allowOptional,
	}, nil
}

func (e *TelemetryEncoder[T]) RecordSize() int { return e.recordSize }

func (e *TelemetryEncoder[T]) DecodeInto(buf []byte, out *T) error {
	if out == nil {
		return fmt.Errorf("telemetry encoder: out is nil")
	}
	if len(buf) < e.recordSize {
		return fmt.Errorf("telemetry encoder: buffer too short (need %d, have %d)", e.recordSize, len(buf))
	}

	v := reflect.ValueOf(out).Elem()
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return fmt.Errorf("telemetry encoder: out must point to a struct")
	}

	for _, b := range e.binds {
		end := b.offset + b.size
		if b.offset < 0 || end > len(buf) {
			return fmt.Errorf("telemetry encoder: buffer too short for %q (need [%d:%d], have %d)",
				b.key, b.offset, end, len(buf))
		}

		fv := v.Field(b.fieldIdx)
		if !fv.CanSet() {
			return fmt.Errorf("telemetry encoder: field %d (%q) is not settable (must be exported)", b.fieldIdx, b.key)
		}

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
			u := e.order.Uint32(buf[b.offset : b.offset+4])
			fv.SetFloat(float64(math.Float32frombits(u)))

		default:
			return fmt.Errorf("telemetry encoder: unsupported primitive %v for %q", b.prim, b.key)
		}
	}
	return nil
}

func (e *TelemetryEncoder[T]) Decode(buf []byte) (T, error) {
	var out T
	if err := e.DecodeInto(buf, &out); err != nil {
		return out, err
	}
	return out, nil
}

func (e *TelemetryEncoder[T]) EncodeInto(dst []byte, v *T) error {
	if v == nil {
		return fmt.Errorf("telemetry encoder: v is nil")
	}
	if len(dst) < e.recordSize {
		return fmt.Errorf("telemetry encoder: dst too short (need %d, have %d)", e.recordSize, len(dst))
	}

	clear(dst[:e.recordSize])

	rv := reflect.ValueOf(v).Elem()
	if rv.Kind() == reflect.Pointer {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return fmt.Errorf("telemetry encoder: v must point to a struct")
	}

	for _, b := range e.binds {
		end := b.offset + b.size
		if b.offset < 0 || end > len(dst) {
			return fmt.Errorf("telemetry encoder: dst too short for %q (need [%d:%d], have %d)",
				b.key, b.offset, end, len(dst))
		}

		fv := rv.Field(b.fieldIdx)

		if b.isPointer {
			if fv.IsNil() {
				if e.allowOptional {
					// Missing => keep zeros.
					continue
				}
				return fmt.Errorf("telemetry encoder: required field %q is nil", b.key)
			}
			fv = fv.Elem()
		}

		switch b.prim {
		case PrimByte:
			dst[b.offset] = byte(fv.Uint())

		case PrimFloat32:
			u := math.Float32bits(float32(fv.Float()))
			e.order.PutUint32(dst[b.offset:b.offset+4], u)

		default:
			return fmt.Errorf("telemetry encoder: unsupported primitive %v for %q", b.prim, b.key)
		}
	}

	return nil
}

func (e *TelemetryEncoder[T]) Encode(v T) ([]byte, error) {
	out := make([]byte, e.recordSize)
	if err := e.EncodeInto(out, &v); err != nil {
		return nil, err
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
