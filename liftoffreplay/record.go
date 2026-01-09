package liftoffreplay

type ReplayRecordFlat struct {
	PosX float32 `json:"pos_x"`
	PosY float32 `json:"pos_y"`
	PosZ float32 `json:"pos_z"`

	RotX float32 `json:"rot_x"`
	RotY float32 `json:"rot_y"`
	RotZ float32 `json:"rot_z"`
	RotW float32 `json:"rot_w"`

	Throttle float32 `json:"throttle"`
	Yaw      float32 `json:"yaw"`
	Pitch    float32 `json:"pitch"`
	Roll     float32 `json:"roll"`

	TimeSec float32 `json:"time_sec"`
}

type ReplayRecordNested struct {
	Position struct {
		X float32 `json:"x"`
		Y float32 `json:"y"`
		Z float32 `json:"z"`
	} `json:"position"`

	Orientation struct {
		X float32 `json:"x"`
		Y float32 `json:"y"`
		Z float32 `json:"z"`
		W float32 `json:"w"`
	} `json:"orientation"`

	Controls struct {
		Throttle float32 `json:"throttle"`
		Yaw      float32 `json:"yaw"`
		Pitch    float32 `json:"pitch"`
		Roll     float32 `json:"roll"`
	} `json:"controls"`

	TimeSec float32 `json:"time_sec"`
}

type ReplayRecordConstraint interface {
	ReplayRecordFlat | ReplayRecordNested
}

func ToNestedRecord(r ReplayRecordFlat) ReplayRecordNested {
	var out ReplayRecordNested

	out.Position.X = r.PosX
	out.Position.Y = r.PosY
	out.Position.Z = r.PosZ

	out.Orientation.X = r.RotX
	out.Orientation.Y = r.RotY
	out.Orientation.Z = r.RotZ
	out.Orientation.W = r.RotW

	out.Controls.Throttle = r.Throttle
	out.Controls.Yaw = r.Yaw
	out.Controls.Pitch = r.Pitch
	out.Controls.Roll = r.Roll

	out.TimeSec = r.TimeSec

	return out
}
