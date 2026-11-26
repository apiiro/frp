// Copyright 2026 The frp Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package xlog

import (
	stdlog "log"
	"strings"
	"sync/atomic"
)

// NewYamuxLogger returns a yamux logger that logs yamux's [ERR] lines as errors and its [WARN] lines as warnings, so a
// failed keepalive or a write timeout on the session shows at the default log level. yamux also reports a session
// that is closing normally as a failed read or write, and those stay at debug. It is a *log.Logger because that is
// what hashicorp/yamux's Config.Logger takes; frp builds against the fatedier fork, whose Logger interface it also
// satisfies, but programs that import frp as a module get hashicorp/yamux, since replace directives do not carry over.
func NewYamuxLogger(xl *Logger) *stdlog.Logger {
	return NewYamuxSessionLogger(xl, nil)
}

// NewYamuxSessionLogger is NewYamuxLogger that also records the session's first yamux error in sessionErr.
func NewYamuxSessionLogger(xl *Logger, sessionErr *YamuxSessionError) *stdlog.Logger {
	return stdlog.New(yamuxWriter{xl: xl, sessionErr: sessionErr}, "", 0)
}

// YamuxSessionError is the first error yamux logged for a session. yamux closes the session itself, on a failed
// keepalive for example, and the reads that follow see only EOF, so this is what says why the session ended.
type YamuxSessionError struct {
	msg atomic.Pointer[string]
}

func (e *YamuxSessionError) Get() string {
	if msg := e.msg.Load(); msg != nil {
		return *msg
	}
	return ""
}

type yamuxWriter struct {
	xl         *Logger
	sessionErr *YamuxSessionError
}

func (w yamuxWriter) Write(p []byte) (int, error) {
	msg := strings.TrimSpace(string(p))
	switch {
	case isSessionClosing(msg):
		w.xl.Debugf("%s", msg)
	case strings.HasPrefix(msg, "[ERR]"):
		w.xl.Errorf("%s", msg)
		if w.sessionErr != nil {
			cause := strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(msg, "[ERR]"), " yamux:"))
			if endsSession(cause) {
				w.sessionErr.msg.CompareAndSwap(nil, &cause)
			}
		}
	case strings.HasPrefix(msg, "[WARN]"):
		w.xl.Warnf("%s", msg)
	default:
		w.xl.Debugf("%s", msg)
	}
	return len(p), nil
}

// yamux [ERR] lines that end the session, in both hashicorp/yamux v0.1.1 and the fatedier fork. Others, such as SYN
// tracking errors, leave the session running, so they cannot be why it closed.
var sessionEndingErrors = []string{
	"keepalive failed",
	"aborted stream open (destination",
	"Failed to read header",
	"Failed to write header",
	"Failed to copy body",
	"Failed to write body",
	"Invalid protocol version",
	"received protocol error go away",
	"received internal error go away",
	"received unexpected go away",
	"duplicate stream declared",
}

func endsSession(cause string) bool {
	for _, prefix := range sessionEndingErrors {
		if strings.HasPrefix(cause, prefix) {
			return true
		}
	}
	return false
}

func isSessionClosing(msg string) bool {
	return strings.HasSuffix(msg, ": EOF") ||
		strings.Contains(msg, "use of closed network connection") ||
		strings.Contains(msg, "session shutdown")
}
