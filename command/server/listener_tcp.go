package server

import (
	"io"
	"net"
	"time"

	"github.com/hashicorp/vault/vault"
)

func tcpListenerFactory(config map[string]string, _ io.Writer) (net.Listener, map[string]string, vault.ReloadFunc, error) {
	addr, ok := config["address"]
	if !ok {
		addr = "127.0.0.1:8200"
	}

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, nil, nil, err
	}

	ln = tcpKeepAliveListener{ln.(*net.TCPListener)}
	props := map[string]string{"addr": addr}
	return listenerWrapTLS(ln, props, config)
}

// tcpKeepAliveListener sets TCP keep-alive timeouts on accepted
// connections. It's used by ListenAndServe and ListenAndServeTLS so
// dead TCP connections (e.g. closing laptop mid-download) eventually
// go away.
//
// This is copied directly from the Go source code.
type tcpKeepAliveListener struct {
	*net.TCPListener
}

func (ln tcpKeepAliveListener) Accept() (c net.Conn, err error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return
	}
	tc.SetKeepAlive(true)
	tc.SetKeepAlivePeriod(3 * time.Minute)
	return tc, nil
}
