package csv

import (
	"encoding/binary"
	"encoding/csv"
	"errors"
	"io"
	"strconv"
	"strings"

	"github.com/mchekh/liftoff-replay/source"
	"github.com/mchekh/liftoff-replay/telemetry"
)

type csvTelemetryRecord struct {
	Timestamp *float32 `telemetry:"Timestamp"`

	PositionX *float32 `telemetry:"PositionX"`
	PositionY *float32 `telemetry:"PositionY"`
	PositionZ *float32 `telemetry:"PositionZ"`

	AttitudeX *float32 `telemetry:"AttitudeX"`
	AttitudeY *float32 `telemetry:"AttitudeY"`
	AttitudeZ *float32 `telemetry:"AttitudeZ"`
	AttitudeW *float32 `telemetry:"AttitudeW"`

	SpeedX *float32 `telemetry:"SpeedX"`
	SpeedY *float32 `telemetry:"SpeedY"`
	SpeedZ *float32 `telemetry:"SpeedZ"`

	GyroPitch *float32 `telemetry:"GyroPitch"`
	GyroRoll  *float32 `telemetry:"GyroRoll"`
	GyroYaw   *float32 `telemetry:"GyroYaw"`

	InputThrottle *float32 `telemetry:"InputThrottle"`
	InputYaw      *float32 `telemetry:"InputYaw"`
	InputPitch    *float32 `telemetry:"InputPitch"`
	InputRoll     *float32 `telemetry:"InputRoll"`

	BatteryVoltage    *float32 `telemetry:"BatteryVoltage"`
	BatteryPercentage *float32 `telemetry:"BatteryPercentage"`

	MotorRPMCount *byte    `telemetry:"MotorRPMCount"`
	MotorRPM_LF   *float32 `telemetry:"MotorRPM_LF"`
	MotorRPM_RF   *float32 `telemetry:"MotorRPM_RF"`
	MotorRPM_LB   *float32 `telemetry:"MotorRPM_LB"`
	MotorRPM_RB   *float32 `telemetry:"MotorRPM_RB"`
}

func fmtF32(v *float32) (string, bool) {
	if v == nil {
		return "", false
	}
	return strconv.FormatFloat(float64(*v), 'f', -1, 32), true
}

func fmtU8(v *byte) (string, bool) {
	if v == nil {
		return "", false
	}
	return strconv.FormatUint(uint64(*v), 10), true
}

type column struct {
	header string
	alias  string
	get    func(r *csvTelemetryRecord) (string, bool)
}

func canonicalizeAlias(s string) string {
	return strings.TrimSpace(s)
}

func columnsAll() []column {
	return []column{
		{header: "position_x", alias: "PositionX", get: func(r *csvTelemetryRecord) (string, bool) { return fmtF32(r.PositionX) }},
		{header: "position_y", alias: "PositionY", get: func(r *csvTelemetryRecord) (string, bool) { return fmtF32(r.PositionY) }},
		{header: "position_z", alias: "PositionZ", get: func(r *csvTelemetryRecord) (string, bool) { return fmtF32(r.PositionZ) }},

		{header: "attitude_x", alias: "AttitudeX", get: func(r *csvTelemetryRecord) (string, bool) { return fmtF32(r.AttitudeX) }},
		{header: "attitude_y", alias: "AttitudeY", get: func(r *csvTelemetryRecord) (string, bool) { return fmtF32(r.AttitudeY) }},
		{header: "attitude_z", alias: "AttitudeZ", get: func(r *csvTelemetryRecord) (string, bool) { return fmtF32(r.AttitudeZ) }},
		{header: "attitude_w", alias: "AttitudeW", get: func(r *csvTelemetryRecord) (string, bool) { return fmtF32(r.AttitudeW) }},

		{header: "speed_x", alias: "SpeedX", get: func(r *csvTelemetryRecord) (string, bool) { return fmtF32(r.SpeedX) }},
		{header: "speed_y", alias: "SpeedY", get: func(r *csvTelemetryRecord) (string, bool) { return fmtF32(r.SpeedY) }},
		{header: "speed_z", alias: "SpeedZ", get: func(r *csvTelemetryRecord) (string, bool) { return fmtF32(r.SpeedZ) }},

		{header: "gyro_pitch", alias: "GyroPitch", get: func(r *csvTelemetryRecord) (string, bool) { return fmtF32(r.GyroPitch) }},
		{header: "gyro_roll", alias: "GyroRoll", get: func(r *csvTelemetryRecord) (string, bool) { return fmtF32(r.GyroRoll) }},
		{header: "gyro_yaw", alias: "GyroYaw", get: func(r *csvTelemetryRecord) (string, bool) { return fmtF32(r.GyroYaw) }},

		{header: "input_throttle", alias: "InputThrottle", get: func(r *csvTelemetryRecord) (string, bool) { return fmtF32(r.InputThrottle) }},
		{header: "input_yaw", alias: "InputYaw", get: func(r *csvTelemetryRecord) (string, bool) { return fmtF32(r.InputYaw) }},
		{header: "input_pitch", alias: "InputPitch", get: func(r *csvTelemetryRecord) (string, bool) { return fmtF32(r.InputPitch) }},
		{header: "input_roll", alias: "InputRoll", get: func(r *csvTelemetryRecord) (string, bool) { return fmtF32(r.InputRoll) }},

		{header: "battery_voltage", alias: "BatteryVoltage", get: func(r *csvTelemetryRecord) (string, bool) { return fmtF32(r.BatteryVoltage) }},
		{header: "battery_percentage", alias: "BatteryPercentage", get: func(r *csvTelemetryRecord) (string, bool) { return fmtF32(r.BatteryPercentage) }},

		{header: "motor_rpm_count", alias: "MotorRPMCount", get: func(r *csvTelemetryRecord) (string, bool) { return fmtU8(r.MotorRPMCount) }},
		{header: "motor_rpm_lf", alias: "MotorRPM_LF", get: func(r *csvTelemetryRecord) (string, bool) { return fmtF32(r.MotorRPM_LF) }},
		{header: "motor_rpm_rf", alias: "MotorRPM_RF", get: func(r *csvTelemetryRecord) (string, bool) { return fmtF32(r.MotorRPM_RF) }},
		{header: "motor_rpm_lb", alias: "MotorRPM_LB", get: func(r *csvTelemetryRecord) (string, bool) { return fmtF32(r.MotorRPM_LB) }},
		{header: "motor_rpm_rb", alias: "MotorRPM_RB", get: func(r *csvTelemetryRecord) (string, bool) { return fmtF32(r.MotorRPM_RB) }},

		{header: "timestamp", alias: "Timestamp", get: func(r *csvTelemetryRecord) (string, bool) { return fmtF32(r.Timestamp) }},
	}
}

type CSVEncoder struct {
	w *csv.Writer

	cols []column
	row  []string

	reuse        csvTelemetryRecord
	telemetryEnc *telemetry.TelemetryEncoder[csvTelemetryRecord]
}

func NewCSVEncoder(schema *telemetry.TelemetrySchema, out io.Writer) (*CSVEncoder, error) {
	if schema == nil {
		return nil, errors.New("csv encoder: schema is nil")
	}

	enc, err := telemetry.NewTelemetryEncoder[csvTelemetryRecord](schema, binary.LittleEndian, true)
	if err != nil {
		return nil, err
	}

	fields := schema.Fields()
	allowed := make(map[string]struct{}, len(fields))
	for _, field := range fields {
		allowed[canonicalizeAlias(field.Desc.Name)] = struct{}{}
	}

	all := columnsAll()
	cols := make([]column, 0, len(all))
	headers := make([]string, 0, len(all))

	for _, c := range all {
		if _, ok := allowed[canonicalizeAlias(c.alias)]; !ok {
			continue
		}
		cols = append(cols, c)
		headers = append(headers, c.header)
	}

	if len(cols) == 0 {
		return nil, errors.New("csv encoder: no columns matched schema fields")
	}

	w := csv.NewWriter(out)

	if err := w.Write(headers); err != nil {
		return nil, err
	}

	return &CSVEncoder{
		w:            w,
		cols:         cols,
		row:          make([]string, len(cols)),
		telemetryEnc: enc,
	}, nil
}

func (e *CSVEncoder) EncodeAll(src source.TelemetrySource) error {
	if src == nil {
		return errors.New("csv encoder: source is nil")
	}

	defer e.w.Flush()

	for {
		frame, err := src.NextFrame()
		if err != nil {
			if err == io.EOF {
				return e.w.Error()
			}
			return err
		}

		if err := e.telemetryEnc.DecodeInto(frame, &e.reuse); err != nil {
			return err
		}

		for i, c := range e.cols {
			v, ok := c.get(&e.reuse)
			if !ok {
				e.row[i] = ""
				continue
			}
			e.row[i] = v
		}

		if err := e.w.Write(e.row); err != nil {
			return err
		}
	}
}
