package dns

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/air-go/rpc/library/loadbalancer"
	"github.com/air-go/rpc/library/logger/nop"
)

func TestDNSDiscoverer(t *testing.T) {
	// success
	func() {
		ctl := gomock.NewController(t)
		defer ctl.Finish()

		lb := loadbalancer.NewMockLoadBalancer(ctl)
		lb.EXPECT().SetAddrs(gomock.Any()).Times(2).Return(nil)

		dd, err := NewDNSDiscoverer("serviceName", lb, []Node{
			{Host: "www.baidu.com", Port: 80, Network: "ip"},
		},
			WithLogger(nop.Logger),
			WithRefreshWindow(time.Millisecond*100),
		)
		assert.Nil(t, err)

		err = dd.Start(context.Background())
		assert.Nil(t, err)

		time.Sleep(time.Millisecond * 150)
		dd.Stop()
	}()

	// fail
	func() {
		ctl := gomock.NewController(t)
		defer ctl.Finish()

		lb := loadbalancer.NewMockLoadBalancer(ctl)
		// lb.EXPECT().SetAddrs(gomock.Any()).Times(1).Return(nil)

		dd, err := NewDNSDiscoverer("serviceName", lb, []Node{
			{Host: "www.baidu.com1", Port: 80, Network: "ip"},
		})
		assert.Nil(t, err)

		err = dd.Start(context.Background())
		assert.NotNil(t, err)
	}()
}
