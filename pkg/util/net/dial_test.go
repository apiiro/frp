package net

import (
	"context"
	"net"
	"net/http"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/net/websocket"
)

func TestDialHookWebsocketCloseWithStalledWrite(t *testing.T) {
	require := require.New(t)

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(err)
	release := make(chan struct{})
	defer close(release)
	server := &http.Server{Handler: websocket.Handler(func(*websocket.Conn) { <-release })}
	go func() { _ = server.Serve(ln) }()
	defer server.Close()

	raw, err := net.Dial("tcp", ln.Addr().String())
	require.NoError(err)
	defer raw.Close()
	_, conn, err := DialHookWebsocket("tcp", "", "")(context.Background(), raw, ln.Addr().String())
	require.NoError(err)

	var written atomic.Int64
	go func() {
		buf := make([]byte, 64*1024)
		for {
			n, err := conn.Write(buf)
			written.Add(int64(n))
			if err != nil {
				return
			}
		}
	}()

	require.Eventually(func() bool {
		before := written.Load()
		time.Sleep(200 * time.Millisecond)
		return before > 0 && written.Load() == before
	}, 10*time.Second, 10*time.Millisecond, "write never stalled")

	closed := make(chan struct{})
	go func() {
		_ = conn.Close()
		close(closed)
	}()
	select {
	case <-closed:
	case <-time.After(websocketCloseTimeout + 5*time.Second):
		t.Fatal("Close blocked behind the stalled Write")
	}
}
