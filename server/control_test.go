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

package server

import (
	"context"
	"math"
	"net"
	"testing"

	"github.com/stretchr/testify/require"

	v1 "github.com/fatedier/frp/pkg/config/v1"
	"github.com/fatedier/frp/pkg/msg"
)

func TestNewControlPoolCountBoundaries(t *testing.T) {
	for _, tc := range []struct {
		name          string
		poolCount     int
		maxPoolCount  int64
		wantErr       string
		wantPoolCount int
		wantCapacity  int
	}{
		{name: "negative pool count below offset", poolCount: -11, maxPoolCount: 5, wantErr: "invalid pool count"},
		{name: "negative pool count at offset", poolCount: -10, maxPoolCount: 5, wantErr: "invalid pool count"},
		{name: "negative pool count", poolCount: -1, maxPoolCount: 5, wantErr: "invalid pool count"},
		{name: "zero pool count", poolCount: 0, maxPoolCount: 5, wantPoolCount: 0, wantCapacity: 10},
		{name: "pool count capped", poolCount: 10, maxPoolCount: 5, wantPoolCount: 5, wantCapacity: 15},
		{name: "maximum int pool count capped", poolCount: math.MaxInt, maxPoolCount: 5, wantPoolCount: 5, wantCapacity: 15},
		{name: "negative maximum", poolCount: 1, maxPoolCount: -1, wantErr: "invalid max pool count"},
		{name: "maximum int64 with small client pool", poolCount: 1, maxPoolCount: math.MaxInt64, wantPoolCount: 1, wantCapacity: 11},
		{name: "maximum int client and server overflow", poolCount: math.MaxInt, maxPoolCount: math.MaxInt64, wantErr: "cannot safely add"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			conn, peer := net.Pipe()
			defer conn.Close()
			defer peer.Close()
			cfg := &v1.ServerConfig{}
			cfg.Transport.MaxPoolCount = tc.maxPoolCount

			ctl, err := NewControl(context.Background(), nil, nil, nil, nil, nil, conn, false,
				&msg.Login{RunID: "pool-count-run", PoolCount: tc.poolCount}, cfg)
			if tc.wantErr != "" {
				require.Nil(t, ctl)
				require.ErrorContains(t, err, tc.wantErr)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.wantPoolCount, ctl.poolCount)
			require.Equal(t, tc.wantCapacity, cap(ctl.workConnCh))
		})
	}
}
