package telemetry

type TelemetrySource interface {
	NextFrame() ([]byte, error)
	Close() error
}
