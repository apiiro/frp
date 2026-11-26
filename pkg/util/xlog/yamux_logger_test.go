package xlog

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestYamuxLoggerLevels(t *testing.T) {
	cases := []struct {
		line  string
		level string
	}{
		{"[ERR] yamux: keepalive failed: i/o deadline reached", "[E]"},
		{"[ERR] yamux: Failed to write header: i/o timeout", "[E]"},
		{"[WARN] yamux: frame for missing stream: Vsn:0 Type:1", "[W]"},
		{"[ERR] yamux: Failed to read header: EOF", "[D]"},
		{"[ERR] yamux: Failed to write header: write tcp 10.0.0.1:1->10.0.0.2:2: use of closed network connection", "[D]"},
		{"yamux: something informational", "[D]"},
	}
	for _, c := range cases {
		t.Run(c.line, func(t *testing.T) {
			buf := captureLogs(t)
			NewYamuxLogger(New()).Printf("%s", c.line)
			require.Contains(t, buf.String(), c.level+" ")
			require.Contains(t, buf.String(), c.line)
		})
	}
}

func TestYamuxSessionLoggerRecordsTheFirstError(t *testing.T) {
	captureLogs(t)
	var sessionErr YamuxSessionError
	logger := NewYamuxSessionLogger(New(), &sessionErr)

	logger.Printf("[WARN] yamux: frame for missing stream: Vsn:0 Type:1")
	logger.Printf("[ERR] yamux: Failed to read header: EOF")
	logger.Printf("[ERR] yamux: SYN tracking out of sync")
	logger.Printf("[ERR] yamux: established stream without inflight SYN (no tracking entry)")
	require.Empty(t, sessionErr.Get(), "a warning, a normal close or an error the session survives is not a cause")

	logger.Printf("[ERR] yamux: keepalive failed: i/o deadline reached")
	logger.Printf("[ERR] yamux: Failed to write header: i/o timeout")
	require.Equal(t, "keepalive failed: i/o deadline reached", sessionErr.Get())
}
