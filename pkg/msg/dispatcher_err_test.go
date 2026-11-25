package msg

import (
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestDispatcherKeepsReadError(t *testing.T) {
	client, server := net.Pipe()
	d := NewDispatcher(client)
	d.Run()

	server.Close()

	select {
	case <-d.Done():
	case <-time.After(time.Second):
		t.Fatal("dispatcher did not stop after the peer closed")
	}
	require.Error(t, d.Err())
}
