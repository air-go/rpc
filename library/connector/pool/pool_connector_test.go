package pool

import (
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConnector(t *testing.T) {
	c, _ := NewConnector("serviceName")
	addr, err := net.ResolveTCPAddr("tcp", "www.baidu.com:80")
	conn, err := c.Connect(context.Background(), addr)
	assert.Nil(t, err)
	c.Back(context.Background(), conn)
}
