package source

type TelemetrySource interface {
	NextFrame() ([]byte, error)
}
