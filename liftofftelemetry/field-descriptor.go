package telemetry

type Primitive uint8

const (
	PrimFloat32 Primitive = iota
	PrimInt32
	PrimByte
)

func (p Primitive) Size() (int, bool) {
	switch p {
	case PrimFloat32, PrimInt32:
		return 4, true
	case PrimByte:
		return 1, true
	default:
		return 0, false
	}
}

type FieldDescriptor struct {
	Name  string
	Alias string
	Type  Primitive
}

func (f FieldDescriptor) Size() (int, bool) { return f.Type.Size() }
