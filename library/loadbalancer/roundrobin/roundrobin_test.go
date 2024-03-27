package roundrobin

import (
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/air-go/rpc/library/addr"
)

func TestRoundRobin(t *testing.T) {
	lb := New()
	_ = lb.SetAddrs([]net.Addr{
		&addr.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 80},
		&addr.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 81},
		&addr.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 82},
		&addr.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 83},
	})
	addr, _ := lb.Pick(context.TODO())
	assert.Equal(t, "127.0.0.1:80", addr.String())
	addr, _ = lb.Pick(context.TODO())
	assert.Equal(t, "127.0.0.1:81", addr.String())
	addr, _ = lb.Pick(context.TODO())
	assert.Equal(t, "127.0.0.1:82", addr.String())
	addr, _ = lb.Pick(context.TODO())
	assert.Equal(t, "127.0.0.1:83", addr.String())
	addr, _ = lb.Pick(context.TODO())
	assert.Equal(t, "127.0.0.1:80", addr.String())
}
