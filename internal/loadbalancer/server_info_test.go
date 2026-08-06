package loadbalancer

import (
	"context"
	"errors"
	"io"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestServerInfoDialRefreshesOfflineBackendsBeforeSelecting(t *testing.T) {
	fast := startStatusBackend(t, 10*time.Millisecond)
	slow := startStatusBackend(t, 80*time.Millisecond)

	fastBackend := NewBackend(fast.addr(), 0, 10)
	slowBackend := NewBackend(slow.addr(), 0, 10)
	fastBackend.SetHealthy(false)
	slowBackend.SetHealthy(false)

	server := NewServerInfo(
		"survival",
		[]*Backend{fastBackend, slowBackend},
		&HealthScoreStrategy{},
		0,
		time.Second,
		time.Second,
		1,
		nil,
	)

	conn, err := server.Dial(context.Background(), nil)
	if err != nil {
		t.Fatalf("Dial() returned an error with reachable offline backends: %v", err)
	}
	defer conn.Close()

	waitForAcceptedConnections(t, fast, 2)
	if got := slow.accepted.Load(); got != 1 {
		t.Fatalf("slow backend accepted %d connections, want 1 health check", got)
	}
	if !fastBackend.IsHealthy() || !slowBackend.IsHealthy() {
		t.Fatal("Dial() did not refresh every reachable backend before selection")
	}
}

func TestServerInfoDialReportsUnavailableAfterRefreshingEveryBackend(t *testing.T) {
	first := startStatusBackend(t, time.Second)
	second := startStatusBackend(t, time.Second)
	backends := []*Backend{
		NewBackend(first.addr(), 0, 10),
		NewBackend(second.addr(), 0, 10),
	}
	for _, backend := range backends {
		backend.SetHealthy(false)
	}

	server := NewServerInfo(
		"survival",
		backends,
		&HealthScoreStrategy{},
		0,
		time.Second,
		20*time.Millisecond,
		1,
		nil,
	)

	_, err := server.Dial(context.Background(), nil)
	if err == nil || !strings.Contains(err.Error(), "no available backend for server survival") {
		t.Fatalf("Dial() error = %v, want no available backend", err)
	}
	waitForAcceptedConnections(t, first, 1)
	waitForAcceptedConnections(t, second, 1)
	for _, backend := range backends {
		if backend.IsHealthy() {
			t.Fatalf("unreachable backend %s was marked healthy", backend.Addr)
		}
	}
}

type statusBackend struct {
	listener net.Listener
	delay    time.Duration
	accepted atomic.Int32
	connsMu  sync.Mutex
	conns    []net.Conn
}

func startStatusBackend(t *testing.T, delay time.Duration) *statusBackend {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	backend := &statusBackend{listener: listener, delay: delay}
	t.Cleanup(func() {
		_ = listener.Close()
		backend.connsMu.Lock()
		defer backend.connsMu.Unlock()
		for _, conn := range backend.conns {
			_ = conn.Close()
		}
	})

	go backend.serve()
	return backend
}

func (b *statusBackend) addr() string {
	return b.listener.Addr().String()
}

func (b *statusBackend) serve() {
	for {
		conn, err := b.listener.Accept()
		if err != nil {
			return
		}
		b.accepted.Add(1)
		b.connsMu.Lock()
		b.conns = append(b.conns, conn)
		b.connsMu.Unlock()
		go b.handle(conn)
	}
}

func (b *statusBackend) handle(conn net.Conn) {
	if err := discardTestPacket(conn); err != nil {
		return
	}
	if err := discardTestPacket(conn); err != nil {
		return
	}
	time.Sleep(b.delay)
	_, _ = conn.Write([]byte{4, 0, 2, '{', '}'})
}

func discardTestPacket(r io.Reader) error {
	length, err := readTestVarInt(r)
	if err != nil {
		return err
	}
	_, err = io.CopyN(io.Discard, r, int64(length))
	return err
}

func readTestVarInt(r io.Reader) (int32, error) {
	var value int32
	for shift := 0; shift < 32; shift += 7 {
		var buf [1]byte
		if _, err := io.ReadFull(r, buf[:]); err != nil {
			return 0, err
		}
		value |= int32(buf[0]&0x7f) << shift
		if buf[0]&0x80 == 0 {
			return value, nil
		}
	}
	return 0, errors.New("VarInt too big")
}

func waitForAcceptedConnections(t *testing.T, backend *statusBackend, want int32) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for backend.accepted.Load() < want && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if got := backend.accepted.Load(); got != want {
		t.Fatalf("backend accepted %d connections, want %d", got, want)
	}
}
