package proxy

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"

	golog "github.com/fatedier/golib/log"
	"github.com/stretchr/testify/require"

	"github.com/fatedier/frp/pkg/util/log"
	"github.com/fatedier/frp/pkg/util/xlog"
)

func TestWorkConnFailuresLogsOncePerWindow(t *testing.T) {
	var buf bytes.Buffer
	old := log.Logger
	log.Logger = golog.New(golog.WithOutput(&buf), golog.WithLevel(golog.TraceLevel))
	t.Cleanup(func() { log.Logger = old })

	now := time.Date(2026, 11, 24, 14, 0, 0, 0, time.UTC)
	f := &workConnFailures{now: func() time.Time { return now }}
	xl := xlog.New()
	timeout := errors.New("timeout trying to get work connection")

	for range 100 {
		f.record(xl, timeout)
		now = now.Add(100 * time.Millisecond)
	}
	require.Equal(t, 1, strings.Count(buf.String(), "\n"), buf.String())

	now = now.Add(time.Minute)
	f.record(xl, timeout)
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	require.Len(t, lines, 2)
	require.Contains(t, lines[1], "failed to get work connection: timeout trying to get work connection (and 99 more times")

	f.record(xl, timeout)
	f.close(xl)
	require.Contains(t, buf.String(), "1 more times")

	buf.Reset()
	f.record(xl, timeout)
	require.Contains(t, buf.String(), "[D]")
	require.Contains(t, buf.String(), "after the proxy closed")
}
