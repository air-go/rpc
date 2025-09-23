package servicer

import (
	"net"
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
)

func TestNode(t *testing.T) {
	convey.Convey("TestNewNode", t, func() {
		convey.Convey("success", func() {
			address := "127.0.0.1:80"
			weight := 10
			addr, _ := net.ResolveTCPAddr("tcp", address)
			node := NewNode(addr, WithWeight(weight))
			assert.Equal(t, node.Addr().String(), address)
			assert.Equal(t, node.Weight(), weight)

			host, port, err := net.SplitHostPort(node.Addr().String())
			assert.Nil(t, err)
			assert.Equal(t, "127.0.0.1", host)
			assert.Equal(t, "80", port)
		})
	})
}
