package net

import (
	"context"
	"net"
	"net/url"
	"time"

	libnet "github.com/fatedier/golib/net"
	"golang.org/x/net/websocket"
)

func DialHookCustomTLSHeadByte(enableTLS bool, disableCustomTLSHeadByte bool) libnet.AfterHookFunc {
	return func(ctx context.Context, c net.Conn, addr string) (context.Context, net.Conn, error) {
		if enableTLS && !disableCustomTLSHeadByte {
			_, err := c.Write([]byte{byte(FRPTLSHeadByte)})
			if err != nil {
				return nil, nil, err
			}
		}
		return ctx, c, nil
	}
}

func DialHookWebsocket(protocol string, host string, auth string) libnet.AfterHookFunc {
	return func(ctx context.Context, c net.Conn, addr string) (context.Context, net.Conn, error) {
		if protocol != "wss" {
			protocol = "ws"
		}
		if host == "" {
			host = addr
		}
		addr = protocol + "://" + host + FrpWebsocketPath
		uri, err := url.Parse(addr)
		if err != nil {
			return nil, nil, err
		}

		origin := "http://" + uri.Host
		cfg, err := websocket.NewConfig(addr, origin)
		if err != nil {
			return nil, nil, err
		}

		if auth != "" {
			cfg.Header.Add("Authorization", auth)
		}

		conn, err := websocket.NewClient(cfg, c)
		if err != nil {
			return nil, nil, err
		}
		return ctx, &websocketClientConn{Conn: conn}, nil
	}
}

const websocketCloseTimeout = 5 * time.Second

type websocketClientConn struct {
	*websocket.Conn
}

// Close bounds the close frame write. x/net's Close waits for the same lock as
// Write, so a Write stalled on a peer that stopped reading would block it, and
// with it yamux's Session.Close and frpc's reconnect, forever.
func (c *websocketClientConn) Close() error {
	_ = c.SetWriteDeadline(time.Now().Add(websocketCloseTimeout))
	return c.Conn.Close()
}
