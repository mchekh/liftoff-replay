package telemetry

import (
	"fmt"
	"strings"
)

var fieldMap = map[string]FieldDescriptor{
	"Timestamp": {Name: "Timestamp", Type: PrimFloat32, Alias: "timestamp"},

	"PositionX": {Name: "PositionX", Type: PrimFloat32, Alias: "position_x"},
	"PositionY": {Name: "PositionY", Type: PrimFloat32, Alias: "position_y"},
	"PositionZ": {Name: "PositionZ", Type: PrimFloat32, Alias: "position_z"},

	"AttitudeX": {Name: "AttitudeX", Type: PrimFloat32, Alias: "attitude_x"},
	"AttitudeY": {Name: "AttitudeY", Type: PrimFloat32, Alias: "attitude_y"},
	"AttitudeZ": {Name: "AttitudeZ", Type: PrimFloat32, Alias: "attitude_z"},
	"AttitudeW": {Name: "AttitudeW", Type: PrimFloat32, Alias: "attitude_w"},

	"SpeedX": {Name: "SpeedX", Type: PrimFloat32, Alias: "speed_x"},
	"SpeedY": {Name: "SpeedY", Type: PrimFloat32, Alias: "speed_y"},
	"SpeedZ": {Name: "SpeedZ", Type: PrimFloat32, Alias: "speed_z"},

	"GyroPitch": {Name: "GyroPitch", Type: PrimFloat32, Alias: "gyro_pitch"},
	"GyroRoll":  {Name: "GyroRoll", Type: PrimFloat32, Alias: "gyro_roll"},
	"GyroYaw":   {Name: "GyroYaw", Type: PrimFloat32, Alias: "gyro_yaw"},

	"InputThrottle": {Name: "InputThrottle", Type: PrimFloat32, Alias: "input_throttle"},
	"InputYaw":      {Name: "InputYaw", Type: PrimFloat32, Alias: "input_yaw"},
	"InputPitch":    {Name: "InputPitch", Type: PrimFloat32, Alias: "input_pitch"},
	"InputRoll":     {Name: "InputRoll", Type: PrimFloat32, Alias: "input_roll"},

	"BatteryVoltage":    {Name: "BatteryVoltage", Type: PrimFloat32, Alias: "battery_voltage"},
	"BatteryPercentage": {Name: "BatteryPercentage", Type: PrimFloat32, Alias: "battery_percentage"},

	"MotorRPMCount": {Name: "MotorRPMCount", Type: PrimByte, Alias: "motor_rpm_count"},
	"MotorRPM_LF":   {Name: "MotorRPM_LF", Type: PrimFloat32, Alias: "motor_rpm_lf"},
	"MotorRPM_RF":   {Name: "MotorRPM_RF", Type: PrimFloat32, Alias: "motor_rpm_rf"},
	"MotorRPM_LB":   {Name: "MotorRPM_LB", Type: PrimFloat32, Alias: "motor_rpm_lb"},
	"MotorRPM_RB":   {Name: "MotorRPM_RB", Type: PrimFloat32, Alias: "motor_rpm_rb"},
}

var compositeFields = map[string][]string{
	"Position": {"PositionX", "PositionY", "PositionZ"},
	"Attitude": {"AttitudeX", "AttitudeY", "AttitudeZ", "AttitudeW"},
	"Velocity": {"SpeedX", "SpeedY", "SpeedZ"},
	"Gyro":     {"GyroPitch", "GyroRoll", "GyroYaw"},
	"Input":    {"InputThrottle", "InputYaw", "InputPitch", "InputRoll"},
	"Battery":  {"BatteryVoltage", "BatteryPercentage"},
	"MotorRPM": {"MotorRPMCount", "MotorRPM_LF", "MotorRPM_RF", "MotorRPM_LB", "MotorRPM_RB"},
}

type FieldInstance struct {
	Desc   FieldDescriptor
	Offset int
}

type TelemetrySchema struct {
	fields []FieldInstance
	byKey  map[string]int
	size   int
}

func SchemaFromStreamFormat(streamFormat []string) (*TelemetrySchema, error) {
	expanded := make([]FieldDescriptor, 0, len(streamFormat)*3)

	addDesc := func(name string) error {
		d, ok := fieldMap[name]
		if !ok {
			return fmt.Errorf("unknown field %q", name)
		}
		expanded = append(expanded, d)
		return nil
	}

	for _, raw := range streamFormat {
		k := strings.TrimSpace(raw)
		if k == "" {
			continue
		}

		if parts, isComposite := compositeFields[k]; isComposite {
			for _, child := range parts {
				if err := addDesc(child); err != nil {
					return nil, fmt.Errorf("composite %q: %w", k, err)
				}
			}
			continue
		}

		if err := addDesc(k); err != nil {
			return nil, fmt.Errorf("streamFormat key %q: %w", k, err)
		}
	}

	s := &TelemetrySchema{
		fields: make([]FieldInstance, 0, len(expanded)),
		byKey:  make(map[string]int, len(expanded)*2),
	}

	offset := 0
	for _, d := range expanded {
		n, ok := d.Size()
		if !ok {
			return nil, fmt.Errorf("unknown size for field %q (type %v)", d.Name, d.Type)
		}

		if _, dup := s.byKey[d.Name]; dup {
			return nil, fmt.Errorf("duplicate field in schema: %q", d.Name)
		}

		idx := len(s.fields)
		fi := FieldInstance{Desc: d, Offset: offset}
		s.fields = append(s.fields, fi)

		s.byKey[d.Name] = idx

		offset += n
	}

	s.size = offset
	return s, nil
}

func (s *TelemetrySchema) Size() int { return s.size }

func (s *TelemetrySchema) Fields() []FieldInstance {
	return s.fields
}

func (s *TelemetrySchema) Field(key string) (FieldInstance, bool) {
	i, ok := s.byKey[key]
	if !ok {
		return FieldInstance{}, false
	}
	return s.fields[i], true
}
