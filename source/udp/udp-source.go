package udp

import (
	"errors"
	"io"
	"net"
	"sync"
	"sync/atomic"
)

type Result struct {
	Frame []byte
	Err   error
}

type UDPConfig struct {
	Addr            string
	MaxDatagramSize int
	QueueSize       int
	ReadBufferBytes int
}

type UdpTelemetrySource struct {
	conn *net.UDPConn
	buff []byte

	results chan Result

	done chan struct{}
	wg   sync.WaitGroup
	once sync.Once

	lossCount atomic.Int64
}

func NewUdpTelemetrySource(cfg UDPConfig) (*UdpTelemetrySource, error) {
	if cfg.Addr == "" {
		return nil, errors.New("udp source: Addr is required")
	}
	if cfg.MaxDatagramSize <= 0 {
		return nil, errors.New("udp source: MaxDatagramSize must be > 0")
	}
	if cfg.QueueSize <= 0 {
		cfg.QueueSize = 1
	}

	udpAddr, err := net.ResolveUDPAddr("udp", cfg.Addr)
	if err != nil {
		return nil, err
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return nil, err
	}

	cleanup := func(e error) (*UdpTelemetrySource, error) {
		_ = conn.Close()
		return nil, e
	}

	if cfg.ReadBufferBytes > 0 {
		if err := conn.SetReadBuffer(cfg.ReadBufferBytes); err != nil {
			return cleanup(err)
		}
	}

	s := &UdpTelemetrySource{
		conn:    conn,
		buff:    make([]byte, cfg.MaxDatagramSize),
		results: make(chan Result, cfg.QueueSize),
		done:    make(chan struct{}),
	}

	s.start()
	return s, nil
}

func (s *UdpTelemetrySource) start() {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		defer close(s.results)

		for {
			n, _, err := s.conn.ReadFromUDP(s.buff)
			if err != nil {

				if errors.Is(err, net.ErrClosed) || s.isDone() {
					return
				}

				s.trySend(Result{Err: err})
				return
			}

			frame := make([]byte, n)
			copy(frame, s.buff[:n])

			if !s.trySend(Result{Frame: frame}) {
				s.lossCount.Add(1)
			}
		}
	}()
}

func (s *UdpTelemetrySource) trySend(r Result) bool {
	select {
	case s.results <- r:
		return true
	case <-s.done:
		return false
	default:
		return false
	}
}

func (s *UdpTelemetrySource) isDone() bool {
	select {
	case <-s.done:
		return true
	default:
		return false
	}
}

func (s *UdpTelemetrySource) NextFrame() ([]byte, error) {
	r, ok := <-s.results
	if !ok {
		return nil, io.EOF
	}
	if r.Err != nil {
		return nil, r.Err
	}
	return r.Frame, nil
}

func (s *UdpTelemetrySource) LossCount() int64 {
	return s.lossCount.Load()
}

func (s *UdpTelemetrySource) Close() error {
	var err error
	s.once.Do(func() {
		close(s.done)
		err = s.conn.Close()
		s.wg.Wait()
	})
	return err
}
