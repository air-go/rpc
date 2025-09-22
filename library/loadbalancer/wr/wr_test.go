// wr is Weighted random
package wr

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"

	"github.com/air-go/rpc/library/servicer"
)

func TestSelector_GetNodes(t *testing.T) {
	s := New()

	addr1, _ := net.ResolveTCPAddr("tcp", "127.0.0.1:80")
	addr2, _ := net.ResolveTCPAddr("tcp", "127.0.0.2:80")
	addr3, _ := net.ResolveTCPAddr("tcp", "127.0.0.3:80")

	node1 := servicer.NewNode(addr1)
	node2 := servicer.NewNode(addr2)
	node3 := servicer.NewNode(addr3)

	nodes := []servicer.Node{
		node1,
		node2,
		node3,
	}
	_ = s.SetNodes(nodes)

	n := s.GetNodes()
	assert.Equal(t, 3, len(n))
}

func TestWR(t *testing.T) {
	convey.Convey("TestWR", t, func() {
		convey.Convey("testNoDeleteHandle same weight", func() {
			addr1, _ := net.ResolveTCPAddr("tcp", "127.0.0.1:80")
			addr2, _ := net.ResolveTCPAddr("tcp", "127.0.0.2:80")
			addr3, _ := net.ResolveTCPAddr("tcp", "127.0.0.3:80")

			node1 := servicer.NewNode(addr1)
			node2 := servicer.NewNode(addr2)
			node3 := servicer.NewNode(addr3)

			nodes := []servicer.Node{
				node1,
				node2,
				node3,
			}
			res := testNoDeleteHandle(t, nodes)
			fmt.Println("\ntestNoDeleteHandle same weight")
			for _, n := range res {
				fmt.Println(n.Addr().String(), ":", n.Statistics())
			}
		})
		convey.Convey("testNoDeleteHandle diff weight", func() {
			addr1, _ := net.ResolveTCPAddr("tcp", "127.0.0.1:80")
			addr2, _ := net.ResolveTCPAddr("tcp", "127.0.0.2:80")
			addr3, _ := net.ResolveTCPAddr("tcp", "127.0.0.3:80")

			node1 := servicer.NewNode(addr1, servicer.WithWeight(2))
			node2 := servicer.NewNode(addr2, servicer.WithWeight(2))
			node3 := servicer.NewNode(addr3, servicer.WithWeight(1))

			nodes := []servicer.Node{
				node1,
				node2,
				node3,
			}

			res := testNoDeleteHandle(t, nodes)
			fmt.Println("\ntestNoDeleteHandle diff weight")
			for _, n := range res {
				fmt.Println(n.Addr().String(), ":", n.Statistics())
			}
		})
		convey.Convey("testDeleteHandle same weight", func() {
			addr1, _ := net.ResolveTCPAddr("tcp", "127.0.0.1:80")
			addr2, _ := net.ResolveTCPAddr("tcp", "127.0.0.2:80")
			addr3, _ := net.ResolveTCPAddr("tcp", "127.0.0.3:80")

			node1 := servicer.NewNode(addr1)
			node2 := servicer.NewNode(addr2)
			node3 := servicer.NewNode(addr3)

			nodes := []servicer.Node{
				node1,
				node2,
				node3,
			}
			res := testDeleteHandle(t, nodes)
			fmt.Println("\ntestDeleteHandle same weight")
			for _, n := range res {
				fmt.Println(n.Addr().String(), ":", n.Statistics())
			}
		})
		convey.Convey("testDeleteHandle diff weight", func() {
			addr1, _ := net.ResolveTCPAddr("tcp", "127.0.0.1:80")
			addr2, _ := net.ResolveTCPAddr("tcp", "127.0.0.2:80")
			addr3, _ := net.ResolveTCPAddr("tcp", "127.0.0.3:80")

			node1 := servicer.NewNode(addr1, servicer.WithWeight(2))
			node2 := servicer.NewNode(addr2, servicer.WithWeight(2))
			node3 := servicer.NewNode(addr3, servicer.WithWeight(1))

			nodes := []servicer.Node{
				node1,
				node2,
				node3,
			}
			res := testDeleteHandle(t, nodes)
			fmt.Println("\ntestDeleteHandle diff weight")
			for _, n := range res {
				fmt.Println(n.Addr().String(), ":", n.Statistics())
			}
		})
	})
}

func testNoDeleteHandle(t *testing.T, nodes []servicer.Node) []servicer.Node {
	s := New()
	_ = s.SetNodes(nodes)

	i := 1
	for {
		if i > 10000 {
			break
		}
		node, _ := s.Pick(context.Background())

		random := rand.Intn(100)
		err := errors.New("error")
		if random != 0 {
			err = nil
		}
		s.Back(node, err)
		i++
	}

	res := s.GetNodes()
	return res
}

func testDeleteHandle(t *testing.T, nodes []servicer.Node) []servicer.Node {
	s := New()
	_ = s.SetNodes(nodes)

	i := 1
	for {
		if i > 9000 {
			break
		}
		node, _ := s.Pick(context.Background())

		random := rand.Intn(100)
		err := errors.New("error")
		if random != 0 {
			err = nil
		}

		s.Back(node, err)
		i++
	}

	del := nodes[2]
	s.deleteNode(del)
	i = 1
	for {
		if i > 1000 {
			break
		}
		node, _ := s.Pick(context.Background())

		random := rand.Intn(10)
		err := errors.New("error")
		if random != 0 {
			err = nil
		}

		assert.Equal(t, node.Addr().String() != del.Addr().String(), true)
		s.Back(node, err)
		i++
	}

	res := s.GetNodes()
	return res
}
