package connections

import (
	"bytes"
	"fmt"
	"net"
	"sync"
)

// BufferedConn wraps a net.Conn and allows for an initial buffer of data to be read before reading from the underlying connection
type BufferedConn struct {
	net.Conn
	buffer   []byte
	bufIndex int
}

func NewBufferedConn(conn net.Conn, buffer []byte) *BufferedConn {
	return &BufferedConn{
		Conn:     conn,
		buffer:   buffer,
		bufIndex: 0,
	}
}

func (bc *BufferedConn) Read(p []byte) (n int, err error) {
	if bc.bufIndex < len(bc.buffer) {
		n = copy(p, bc.buffer[bc.bufIndex:])
		bc.bufIndex += n
		if bc.bufIndex >= len(bc.buffer) {
			bc.buffer = nil
			bc.bufIndex = 0
		}
		if n == len(p) {
			return n, nil
		}
		m, err := bc.Conn.Read(p[n:])
		return n + m, err
	}
	return bc.Conn.Read(p)
}

// ReplayConn wraps a net.Conn and replays a predefined byte slice on the first read before reading from the underlying connection
type ReplayConn struct {
	net.Conn
	replayData []byte
	replayed   bool
	mu         sync.Mutex
}

func NewReplayConn(conn net.Conn, replayData []byte) *ReplayConn {
	return &ReplayConn{
		Conn:       conn,
		replayData: replayData,
		replayed:   false,
	}
}

func (r *ReplayConn) Read(b []byte) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.replayed && len(r.replayData) > 0 {
		n := copy(b, r.replayData)
		if n < len(r.replayData) {
			r.replayData = r.replayData[n:]
		} else {
			r.replayed = true
			r.replayData = nil
		}
		return n, nil
	}

	return r.Conn.Read(b)
}

// CapturingConn wraps a net.Conn and captures all data read from it into a bytes.Buffer
type CapturingConn struct {
	net.Conn
	capture *bytes.Buffer
}

func NewCapturingConn(conn net.Conn, capture *bytes.Buffer) *CapturingConn {
	return &CapturingConn{
		Conn:    conn,
		capture: capture,
	}
}

func (c *CapturingConn) Read(p []byte) (n int, err error) {
	n, err = c.Conn.Read(p)
	if n > 0 && c.capture != nil {
		c.capture.Write(p[:n])
	}
	return n, err
}

// SingleConnListener is a net.Listener that returns a single connection and then always returns an error on Accept
type SingleConnListener struct {
	conn net.Conn
	used bool
	mu   sync.Mutex
}

func NewSingleConnListener(conn net.Conn) *SingleConnListener {
	return &SingleConnListener{conn: conn}
}

func (l *SingleConnListener) Accept() (net.Conn, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.used {
		return nil, fmt.Errorf("listener already used")
	}
	l.used = true

	if l.conn == nil {
		return nil, fmt.Errorf("no connection available")
	}

	conn := l.conn
	l.conn = nil
	return conn, nil
}

func (l *SingleConnListener) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.conn != nil {
		err := l.conn.Close()
		l.conn = nil
		return err
	}
	return nil
}

func (l *SingleConnListener) Addr() net.Addr {
	return l.conn.LocalAddr()
}
