package dns

import (
	"context"
	"net"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/why444216978/go-util/assert"
	ucontext "github.com/why444216978/go-util/context"
	"github.com/why444216978/go-util/nopanic"

	"github.com/air-go/rpc/library/discoverer"
	"github.com/air-go/rpc/library/loadbalancer"
	"github.com/air-go/rpc/library/logger"
	"github.com/air-go/rpc/library/logger/setup"
	"github.com/air-go/rpc/library/servicer"
)

func defaultOptions() *discoverer.Options {
	return &discoverer.Options{
		RefreshWindow: 3 * time.Second,
	}
}

type dnsDiscoverer struct {
	*discoverer.Options
	setup.SetupLogger
	serviceName string
	lb          loadbalancer.LoadBalancer
	startOnce   sync.Once
	stop        context.CancelFunc
}

var _ discoverer.Discoverer = (*dnsDiscoverer)(nil)

func NewDNSDiscoverer(serviceName string, lb loadbalancer.LoadBalancer, opts ...discoverer.OptionFunc) (*dnsDiscoverer, error) {
	if assert.IsNil(lb) {
		return nil, errors.New("new dns discoverer loadbalancer nil")
	}

	opt := defaultOptions()
	for _, o := range opts {
		o(opt)
	}

	if len(opt.IDCNodes) == 0 {
		return nil, errors.New("new dns discoverer nodes nil")
	}

	dd := &dnsDiscoverer{
		Options:     opt,
		serviceName: serviceName,
		lb:          lb,
	}

	dd.SetupLogger.SetLogger(opt.Logger)

	return dd, nil
}

func (dd *dnsDiscoverer) Start(ctx context.Context) (err error) {
	dd.startOnce.Do(func() {
		ctx, dd.stop = context.WithCancel(ucontext.RemoveDeadline(ctx))

		if err = dd.discover(ctx, false); err != nil {
			return
		}

		dd.loop(ctx)
	})

	return
}

func (dd *dnsDiscoverer) Stop() error {
	dd.stop()
	return nil
}

// discover force get newest and notify loadbalancer to update nodes.
func (dd *dnsDiscoverer) discover(ctx context.Context, allowError bool) error {
	addrs, err := dd.getAddrs(ctx, allowError)
	if err != nil {
		return err
	}
	if len(addrs) == 0 {
		return nil
	}

	if err = dd.lb.SetNodes(addrs); err != nil {
		dd.AutoLogger().Error(ctx, "dnsDiscoverSetAddressesErr",
			logger.Reflect(logger.ServiceName, dd.serviceName),
			logger.Error(err),
		)
		return err
	}

	return nil
}

func (dd *dnsDiscoverer) getAddrs(ctx context.Context, allowError bool) ([]servicer.Node, error) {
	addrs := []servicer.Node{}
	for _, n := range dd.Options.IDCNodes {
		ips, err := net.DefaultResolver.LookupIPAddr(ctx, n.Host)
		if err != nil {
			err = asDNSError(err)
			dd.AutoLogger().Error(ctx, "dnsDiscoverLookupIPErr",
				logger.Reflect(logger.ServiceName, dd.serviceName),
				logger.Error(err),
			)
			if allowError {
				continue
			}
			return addrs, err
		}
		for _, ip := range ips {
			addrs = append(addrs, servicer.NewNode(&net.TCPAddr{
				IP:   ip.IP,
				Port: n.Port,
				Zone: ip.Zone,
			}))
		}
	}
	return addrs, nil
}

func (dd *dnsDiscoverer) loop(ctx context.Context) {
	go nopanic.GoVoid(ctx, func() {
		timer := time.NewTicker(dd.RefreshWindow)
		defer timer.Stop()

		for {
			select {
			case <-ctx.Done():
				dd.AutoLogger().Info(ctx, "discoverSetAddressesErr",
					logger.Reflect(logger.ServiceName, dd.serviceName),
					logger.Error(ctx.Err()),
				)
				return
			case <-timer.C:
				_ = dd.discover(ctx, true)
			}
		}
	})
}

func asDNSError(err error) *net.DNSError {
	var de *net.DNSError
	errors.As(err, &de)
	// de.IsNotFound
	return de
}
