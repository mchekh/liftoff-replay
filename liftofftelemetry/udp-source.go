package telemetry

import "net"

type UdpTelemetrySource struct {
	conn *net.UDPConn
	buff []byte
}

func NewUdpTelemetrySource(conn *net.UDPConn, maxDatagramSize int) *UdpTelemetrySource {
	return &UdpTelemetrySource{
		conn: conn,
		buff: make([]byte, maxDatagramSize),
	}

}

func (s *UdpTelemetrySource) NextFrame() ([]byte, error) {
	n, _, err := s.conn.ReadFromUDP(s.buff)
	if err != nil {
		return nil, err
	}

	return s.buff[:n], nil

}

func (s *UdpTelemetrySource) Close() error {
	return s.conn.Close()
}
