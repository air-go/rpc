// wr is Weighted random
package wr

import (
	"errors"
	"fmt"
	"math/rand"
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"

	"github.com/air-go/rpc/library/selector"
	"github.com/air-go/rpc/library/servicer"
)

func TestNewNode(t *testing.T) {
	convey.Convey("TestNewNode", t, func() {
		convey.Convey("success", func() {
			ip := "127.0.0.1"
			port := 80
			weight := 10
			node := servicer.NewNode(ip, port, servicer.WithWeight(weight))
			assert.Equal(t, node.Address(), servicer.GenerateAddress(ip, port))
			assert.Equal(t, node.Weight(), weight)
		})
	})
}

func TestNode_Address(t *testing.T) {
}

func TestNode_Meta(t *testing.T) {
}

func TestNode_Statistics(t *testing.T) {
}

func TestNode_Weight(t *testing.T) {
}

func TestNode_IncrSuccess(t *testing.T) {
}

func TestNode_IncrFail(t *testing.T) {
}

func TestWithServiceName(t *testing.T) {
}

func TestNewSelector(t *testing.T) {
}

func TestSelector_ServiceName(t *testing.T) {
	convey.Convey("TestSelector_ServiceName", t, func() {
		convey.Convey("success", func() {
			s := NewSelector("test_service")
			serviceName := s.ServiceName()
			assert.Equal(t, serviceName, "test_service")
		})
	})
}

func TestSelector_AddNode(t *testing.T) {
}

func TestSelector_DeleteNode(t *testing.T) {
}

func TestSelector_GetNodes(t *testing.T) {
}

func TestSelector_Select(t *testing.T) {
}

func TestSelector_AfterHandle(t *testing.T) {
}

func TestSelector_checkSameWeight(t *testing.T) {
}

func TestSelector_sortOffset(t *testing.T) {
}

func TestSelector_node2WRNode(t *testing.T) {
}

func TestWR(t *testing.T) {
	convey.Convey("TestWR", t, func() {
		convey.Convey("testNoDeleteHandle same weight", func() {
			nodes := []servicer.Node{
				servicer.NewNode("127.0.0.1", 80),
				servicer.NewNode("127.0.0.2", 80),
				servicer.NewNode("127.0.0.3", 80),
			}
			res := testNoDeleteHandle(t, nodes)
			fmt.Println("\ntestNoDeleteHandle same weight")
			for _, n := range res {
				fmt.Println(n.Address(), ":", n.Statistics())
			}
		})
		convey.Convey("testNoDeleteHandle diff weight", func() {
			nodes := []servicer.Node{
				servicer.NewNode("127.0.0.1", 80, servicer.WithWeight(2)),
				servicer.NewNode("127.0.0.2", 80, servicer.WithWeight(2)),
				servicer.NewNode("127.0.0.3", 80, servicer.WithWeight(1)),
			}
			res := testNoDeleteHandle(t, nodes)
			fmt.Println("\ntestNoDeleteHandle diff weight")
			for _, n := range res {
				fmt.Println(n.Address(), ":", n.Statistics())
			}
		})
		convey.Convey("testDeleteHandle same weight", func() {
			nodes := []servicer.Node{
				servicer.NewNode("127.0.0.1", 80),
				servicer.NewNode("127.0.0.2", 80),
				servicer.NewNode("127.0.0.3", 80),
			}
			res := testDeleteHandle(t, nodes)
			fmt.Println("\ntestDeleteHandle same weight")
			for _, n := range res {
				fmt.Println(n.Address(), ":", n.Statistics())
			}
		})
		convey.Convey("testDeleteHandle diff weight", func() {
			nodes := []servicer.Node{
				servicer.NewNode("127.0.0.1", 80, servicer.WithWeight(2)),
				servicer.NewNode("127.0.0.2", 80, servicer.WithWeight(2)),
				servicer.NewNode("127.0.0.3", 80, servicer.WithWeight(1)),
			}
			res := testDeleteHandle(t, nodes)
			fmt.Println("\ntestDeleteHandle diff weight")
			for _, n := range res {
				fmt.Println(n.Address(), ":", n.Statistics())
			}
		})
	})
}

func testNoDeleteHandle(t *testing.T, nodes []servicer.Node) []servicer.Node {
	s := NewSelector("test_service")

	for _, node := range nodes {
		s.AddNode(node)
	}

	i := 1
	for {
		if i > 10000 {
			break
		}
		node, _ := s.Select()

		random := rand.Intn(100)
		err := errors.New("error")
		if random != 0 {
			err = nil
		}
		s.AfterHandle(selector.HandleInfo{Node: node, Err: err})
		i++
	}

	res, _ := s.GetNodes()
	return res
}

func testDeleteHandle(t *testing.T, nodes []servicer.Node) []servicer.Node {
	s := NewSelector("test_service")

	for _, node := range nodes {
		s.AddNode(node)
	}

	i := 1
	for {
		if i > 9000 {
			break
		}
		node, _ := s.Select()

		random := rand.Intn(100)
		err := errors.New("error")
		if random != 0 {
			err = nil
		}

		s.AfterHandle(selector.HandleInfo{Node: node, Err: err})
		i++
	}

	del := nodes[2]
	_ = s.DeleteNode(del)
	i = 1
	for {
		if i > 1000 {
			break
		}
		node, _ := s.Select()

		random := rand.Intn(10)
		err := errors.New("error")
		if random != 0 {
			err = nil
		}

		assert.Equal(t, node.Address() != del.Address(), true)
		s.AfterHandle(selector.HandleInfo{Node: node, Err: err})
		i++
	}

	res, _ := s.GetNodes()
	return res
}
