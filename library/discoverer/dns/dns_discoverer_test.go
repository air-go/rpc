package dns

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/air-go/rpc/library/discoverer"
	"github.com/air-go/rpc/library/loadbalancer"
	"github.com/air-go/rpc/library/logger/nop"
)

func TestDNSDiscoverer(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	lb := loadbalancer.NewMockLoadBalancer(ctl)
	lb.EXPECT().SetNodes(gomock.Any()).Times(2).Return(nil)

	dd, err := NewDNSDiscoverer("serviceName", lb,
		discoverer.WithLogger(nop.Logger),
		discoverer.WithRefreshWindow(time.Millisecond*100),
		discoverer.WithIDCNodes([]discoverer.IDCNode{
			{
				IDC:     "bj",
				Network: "tcp",
				Host:    "www.baidu.com",
				Port:    80,
			},
			{
				IDC:     "gz",
				Network: "tcp",
				Host:    "127.0.0.1",
				Port:    80,
			},
		}),
	)
	assert.Nil(t, err)

	err = dd.Start(context.Background())
	assert.Nil(t, err)

	time.Sleep(time.Millisecond * 150)
	dd.Stop()
}
