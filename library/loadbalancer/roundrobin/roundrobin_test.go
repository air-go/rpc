package roundrobin

import (
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/air-go/rpc/library/servicer"
)

func TestRoundRobin(t *testing.T) {
	lb := New()

	addr0, _ := net.ResolveTCPAddr("tcp", "127.0.0.1:80")
	addr1, _ := net.ResolveTCPAddr("tcp", "127.0.0.1:81")
	addr2, _ := net.ResolveTCPAddr("tcp", "127.0.0.1:82")
	addr3, _ := net.ResolveTCPAddr("tcp", "127.0.0.1:83")

	node0 := servicer.NewNode(addr0)
	node1 := servicer.NewNode(addr1)
	node2 := servicer.NewNode(addr2)
	node3 := servicer.NewNode(addr3)

	_ = lb.SetNodes([]servicer.Node{
		node0,
		node1,
		node2,
		node3,
	})
	addr, _ := lb.Pick(context.TODO())
	assert.Equal(t, "127.0.0.1:80", addr.Addr().String())
	addr, _ = lb.Pick(context.TODO())
	assert.Equal(t, "127.0.0.1:81", addr.Addr().String())
	addr, _ = lb.Pick(context.TODO())
	assert.Equal(t, "127.0.0.1:82", addr.Addr().String())
	addr, _ = lb.Pick(context.TODO())
	assert.Equal(t, "127.0.0.1:83", addr.Addr().String())
	addr, _ = lb.Pick(context.TODO())
	assert.Equal(t, "127.0.0.1:80", addr.Addr().String())
}
