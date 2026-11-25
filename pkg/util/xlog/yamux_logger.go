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
	"fmt"
	"strings"
)

// YamuxLogger is a yamux Logger that logs yamux's [ERR] lines as errors and its [WARN] lines as warnings, so a
// failed keepalive or a write timeout on the session shows at the default log level. yamux also reports a session
// that is closing normally as a failed read or write, and those stay at debug.
type YamuxLogger struct {
	xl *Logger
}

func NewYamuxLogger(xl *Logger) YamuxLogger {
	return YamuxLogger{xl: xl}
}

func (l YamuxLogger) Print(v ...any) {
	l.log(fmt.Sprint(v...))
}

func (l YamuxLogger) Printf(format string, v ...any) {
	l.log(fmt.Sprintf(format, v...))
}

func (l YamuxLogger) Println(v ...any) {
	l.log(fmt.Sprintln(v...))
}

func (l YamuxLogger) log(msg string) {
	msg = strings.TrimSpace(msg)
	switch {
	case isSessionClosing(msg):
		l.xl.Debugf("%s", msg)
	case strings.HasPrefix(msg, "[ERR]"):
		l.xl.Errorf("%s", msg)
	case strings.HasPrefix(msg, "[WARN]"):
		l.xl.Warnf("%s", msg)
	default:
		l.xl.Debugf("%s", msg)
	}
}

func isSessionClosing(msg string) bool {
	return strings.HasSuffix(msg, ": EOF") ||
		strings.Contains(msg, "use of closed network connection") ||
		strings.Contains(msg, "session shutdown")
}
