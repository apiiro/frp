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

package proxy

import (
	"sync"
	"time"

	"github.com/fatedier/frp/pkg/util/xlog"
)

const workConnFailureWindow = time.Minute

// workConnFailures logs the first failure to get a work connection in each window and counts the rest, so a client
// that stops answering ReqWorkConn writes one line per minute per proxy instead of one per user connection. The count
// for a window is logged with the next failure after it, or when the proxy closes.
type workConnFailures struct {
	mu          sync.Mutex
	now         func() time.Time
	windowStart time.Time
	suppressed  int
	lastErr     error
}

func (f *workConnFailures) record(xl *xlog.Logger, err error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	now := f.clock()
	if !f.windowStart.IsZero() && now.Sub(f.windowStart) < workConnFailureWindow {
		f.suppressed++
		f.lastErr = err
		return
	}
	f.flushLocked(xl, now)
	f.windowStart = now
	xl.Warnf("failed to get work connection: %v", err)
}

func (f *workConnFailures) flush(xl *xlog.Logger) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.flushLocked(xl, f.clock())
}

func (f *workConnFailures) flushLocked(xl *xlog.Logger, now time.Time) {
	if f.suppressed == 0 {
		return
	}
	xl.Warnf("failed to get work connection %d more times in the %s after the last logged failure, last error: %v",
		f.suppressed, now.Sub(f.windowStart).Round(time.Second), f.lastErr)
	f.suppressed = 0
	f.lastErr = nil
}

func (f *workConnFailures) clock() time.Time {
	if f.now != nil {
		return f.now()
	}
	return time.Now()
}
