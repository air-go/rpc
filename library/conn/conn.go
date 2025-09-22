package conn

import (
	"net"
)

// TODO reserve
type conn struct {
	net.Conn
	tracers []ConnTracer
}

func Wrap(c net.Conn, tracers ...ConnTracer) net.Conn {
	return &conn{
		Conn:    c,
		tracers: tracers,
	}
}

func (c *conn) Read(b []byte) (int, error) {
	for _, t := range c.tracers {
		t.BeforeRead(c, b)
	}

	n, err := c.Read(b)

	for _, t := range c.tracers {
		t.AfterRead(c, b, n, err)
	}

	return n, err
}

func (c *conn) Write(b []byte) (int, error) {
	for _, t := range c.tracers {
		t.BeforeWrite(c, b)
	}

	n, err := c.Write(b)

	for _, t := range c.tracers {
		t.AfterWrite(c, b, n, err)
	}

	return n, err
}

func (c *conn) Close() error {
	for _, t := range c.tracers {
		t.BeforeClose(c)
	}

	err := c.Close()

	for _, t := range c.tracers {
		t.AfterClose(c, err)
	}

	return err
}
