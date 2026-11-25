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
)

// NewYamuxLogger returns a yamux logger that logs yamux's [ERR] lines as errors and its [WARN] lines as warnings, so a
// failed keepalive or a write timeout on the session shows at the default log level. yamux also reports a session
// that is closing normally as a failed read or write, and those stay at debug. It is a *log.Logger because that is
// what hashicorp/yamux's Config.Logger takes; frp builds against the fatedier fork, whose Logger interface it also
// satisfies, but programs that import frp as a module get hashicorp/yamux, since replace directives do not carry over.
func NewYamuxLogger(xl *Logger) *stdlog.Logger {
	return stdlog.New(yamuxWriter{xl: xl}, "", 0)
}

type yamuxWriter struct {
	xl *Logger
}

func (w yamuxWriter) Write(p []byte) (int, error) {
	msg := strings.TrimSpace(string(p))
	switch {
	case isSessionClosing(msg):
		w.xl.Debugf("%s", msg)
	case strings.HasPrefix(msg, "[ERR]"):
		w.xl.Errorf("%s", msg)
	case strings.HasPrefix(msg, "[WARN]"):
		w.xl.Warnf("%s", msg)
	default:
		w.xl.Debugf("%s", msg)
	}
	return len(p), nil
}

func isSessionClosing(msg string) bool {
	return strings.HasSuffix(msg, ": EOF") ||
		strings.Contains(msg, "use of closed network connection") ||
		strings.Contains(msg, "session shutdown")
}
