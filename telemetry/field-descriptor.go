package telemetry

type Primitive uint8

const (
	PrimInvalid Primitive = iota
	PrimFloat32
	PrimByte
)

type FieldDescriptor struct {
	Name  string
	Alias string
	Type  Primitive
}

func (d FieldDescriptor) Size() (int, bool) {
	switch d.Type {
	case PrimFloat32:
		return 4, true
	case PrimByte:
		return 1, true
	default:
		return 0, false
	}
}
