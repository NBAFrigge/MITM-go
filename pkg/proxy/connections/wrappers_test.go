package connections

import (
	"bytes"
	"io"
	"net"
	"testing"
	"time"
)

type mockConn struct {
	readBuf  *bytes.Buffer
	writeBuf *bytes.Buffer
	closed   bool
}

func newMockConn(data []byte) *mockConn {
	return &mockConn{
		readBuf:  bytes.NewBuffer(data),
		writeBuf: &bytes.Buffer{},
	}
}

func (m *mockConn) Read(b []byte) (int, error)         { return m.readBuf.Read(b) }
func (m *mockConn) Write(b []byte) (int, error)        { return m.writeBuf.Write(b) }
func (m *mockConn) Close() error                       { m.closed = true; return nil }
func (m *mockConn) LocalAddr() net.Addr                { return &net.TCPAddr{} }
func (m *mockConn) RemoteAddr() net.Addr               { return &net.TCPAddr{} }
func (m *mockConn) SetDeadline(t time.Time) error      { return nil }
func (m *mockConn) SetReadDeadline(t time.Time) error  { return nil }
func (m *mockConn) SetWriteDeadline(t time.Time) error { return nil }

func TestBufferedConn_ReadsBufferBeforeConn(t *testing.T) {
	underlying := newMockConn([]byte("conn_data"))
	buf := NewBufferedConn(underlying, []byte("buf_data"))

	out := make([]byte, 8)
	n, err := buf.Read(out)
	if err != nil && err != io.EOF {
		t.Fatalf("Read failed: %v", err)
	}
	if string(out[:n]) != "buf_data" {
		t.Errorf("first read = %q, want buf_data", string(out[:n]))
	}
}

func TestBufferedConn_FallsThroughToConn(t *testing.T) {
	underlying := newMockConn([]byte("conn_data"))
	buf := NewBufferedConn(underlying, nil)

	out := make([]byte, 9)
	n, _ := buf.Read(out)
	if string(out[:n]) != "conn_data" {
		t.Errorf("read with nil buffer = %q, want conn_data", string(out[:n]))
	}
}

func TestBufferedConn_FullSequentialRead(t *testing.T) {
	underlying := newMockConn([]byte("world"))
	buf := NewBufferedConn(underlying, []byte("hello"))

	all, err := io.ReadAll(buf)
	if err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}
	if string(all) != "helloworld" {
		t.Errorf("ReadAll = %q, want helloworld", string(all))
	}
}

func TestBufferedConn_BufferClearedAfterExhaustion(t *testing.T) {
	underlying := newMockConn([]byte("b"))
	buf := NewBufferedConn(underlying, []byte("a"))

	io.ReadAll(buf)

	if buf.buffer != nil {
		t.Error("buffer should be nil after exhaustion")
	}
}

func TestReplayConn_ReplaysBeforeConn(t *testing.T) {
	underlying := newMockConn([]byte("live"))
	rc := NewReplayConn(underlying, []byte("replay"))

	out := make([]byte, 6)
	n, _ := rc.Read(out)
	if string(out[:n]) != "replay" {
		t.Errorf("first read = %q, want replay", string(out[:n]))
	}
}

func TestReplayConn_FallsThroughAfterReplay(t *testing.T) {
	underlying := newMockConn([]byte("live"))
	rc := NewReplayConn(underlying, []byte("replay"))

	io.ReadFull(rc, make([]byte, 6))

	out := make([]byte, 4)
	n, _ := rc.Read(out)
	if string(out[:n]) != "live" {
		t.Errorf("post-replay read = %q, want live", string(out[:n]))
	}
}

func TestReplayConn_PartialReadsExhaustReplay(t *testing.T) {
	underlying := newMockConn([]byte("live"))
	rc := NewReplayConn(underlying, []byte("ab"))

	out := make([]byte, 1)
	rc.Read(out)
	if string(out[:1]) != "a" {
		t.Errorf("first partial read = %q, want a", string(out[:1]))
	}
	rc.Read(out)
	if string(out[:1]) != "b" {
		t.Errorf("second partial read = %q, want b", string(out[:1]))
	}

	out2 := make([]byte, 4)
	n, _ := rc.Read(out2)
	if string(out2[:n]) != "live" {
		t.Errorf("post-replay read = %q, want live", string(out2[:n]))
	}
}

func TestReplayConn_ReplayedOnlyOnce(t *testing.T) {
	underlying := newMockConn([]byte("live"))
	rc := NewReplayConn(underlying, []byte("replay"))

	all, _ := io.ReadAll(rc)
	if string(all) != "replaylive" {
		t.Errorf("ReadAll = %q, want replaylive", string(all))
	}
}

func TestCapturingConn_CapturesAllReads(t *testing.T) {
	underlying := newMockConn([]byte("captured data"))
	var capture bytes.Buffer
	conn := NewCapturingConn(underlying, &capture)

	out := make([]byte, 8)
	io.ReadFull(conn, out)

	if capture.String() != "captured" {
		t.Errorf("capture = %q, want captured", capture.String())
	}
}

func TestCapturingConn_NilCapture(t *testing.T) {
	underlying := newMockConn([]byte("data"))
	conn := NewCapturingConn(underlying, nil)

	out := make([]byte, 4)
	_, err := conn.Read(out)
	if err != nil && err != io.EOF {
		t.Fatalf("Read with nil capture failed: %v", err)
	}
}

func TestCapturingConn_CapturesMultipleReads(t *testing.T) {
	underlying := newMockConn([]byte("abcd"))
	var capture bytes.Buffer
	conn := NewCapturingConn(underlying, &capture)

	buf := make([]byte, 2)
	conn.Read(buf)
	conn.Read(buf)

	if capture.String() != "abcd" {
		t.Errorf("capture = %q, want abcd", capture.String())
	}
}

func TestSingleConnListener_AcceptReturnsConn(t *testing.T) {
	underlying := newMockConn(nil)
	listener := NewSingleConnListener(underlying)

	conn, err := listener.Accept()
	if err != nil {
		t.Fatalf("Accept failed: %v", err)
	}
	if conn == nil {
		t.Fatal("Accept returned nil conn")
	}
}

func TestSingleConnListener_AcceptOnlyOnce(t *testing.T) {
	underlying := newMockConn(nil)
	listener := NewSingleConnListener(underlying)

	listener.Accept()

	_, err := listener.Accept()
	if err == nil {
		t.Error("second Accept should return error")
	}
}

func TestSingleConnListener_AddrAfterClose(t *testing.T) {
	underlying := newMockConn(nil)
	listener := NewSingleConnListener(underlying)

	listener.Close()

	addr := listener.Addr()
	if addr == nil {
		t.Error("Addr() should not return nil after Close")
	}
}

func TestSingleConnListener_AddrAfterAccept(t *testing.T) {
	underlying := newMockConn(nil)
	listener := NewSingleConnListener(underlying)

	listener.Accept()

	addr := listener.Addr()
	if addr == nil {
		t.Error("Addr() should not return nil after Accept")
	}
}

func TestSingleConnListener_CloseIsIdempotent(t *testing.T) {
	underlying := newMockConn(nil)
	listener := NewSingleConnListener(underlying)

	if err := listener.Close(); err != nil {
		t.Fatalf("first Close failed: %v", err)
	}
	if err := listener.Close(); err != nil {
		t.Errorf("second Close should not error, got: %v", err)
	}
}

func TestSingleConnListener_ConnNilAfterClose(t *testing.T) {
	underlying := newMockConn(nil)
	listener := NewSingleConnListener(underlying)

	listener.Close()

	if listener.conn != nil {
		t.Error("conn should be nil after Close")
	}
}

func TestSingleConnListener_ConnNilAfterAccept(t *testing.T) {
	underlying := newMockConn(nil)
	listener := NewSingleConnListener(underlying)

	listener.Accept()

	if listener.conn != nil {
		t.Error("conn should be nil after Accept")
	}
}
