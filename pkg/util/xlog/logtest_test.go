package xlog

import (
	"bytes"
	"testing"

	golog "github.com/fatedier/golib/log"

	"github.com/fatedier/frp/pkg/util/log"
)

// captureLogs points frp's logger at a buffer for the duration of the test.
func captureLogs(t *testing.T) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	old := log.Logger
	log.Logger = golog.New(golog.WithOutput(&buf), golog.WithLevel(golog.TraceLevel))
	t.Cleanup(func() { log.Logger = old })
	return &buf
}
