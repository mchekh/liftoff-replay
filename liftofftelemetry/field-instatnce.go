package telemetry

type FieldInstance struct {
	Desc   FieldDescriptor
	Offset int
}

func (fi FieldInstance) Size() (int, bool) { return fi.Desc.Size() }
