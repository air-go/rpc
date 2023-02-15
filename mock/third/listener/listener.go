package listener

import (
	"net"
)

var _ net.Listener = (*Listener)(nil)

type Listener struct {
	net.Listener
}

func (l *Listener) Accept() (net.Conn, error) {
	conn, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}

	nc := &connWrap{
		Conn: conn,
	}
	return nc, nil
}

func (l *Listener) Close() error {
	return l.Listener.Close()
}

func (l *Listener) Addr() net.Addr {
	return l.Listener.Addr()
}

type connWrap struct {
	net.Conn
}

func (c *connWrap) Write(b []byte) (n int, err error) {
	return c.Conn.Write(b)
}
