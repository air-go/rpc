package load

import (
	"context"
	"path/filepath"

	utilDir "github.com/why444216978/go-util/dir"

	"github.com/air-go/rpc/library/config"
	"github.com/air-go/rpc/library/discoverer"
	df "github.com/air-go/rpc/library/discoverer/factory"
	"github.com/air-go/rpc/library/etcd"
	lf "github.com/air-go/rpc/library/loadbalancer/factory"
	"github.com/air-go/rpc/library/servicer"
	"github.com/air-go/rpc/library/servicer/service"
)

func LoadGlobPattern(ctx context.Context, path, suffix string, etcd *etcd.Etcd) (err error) {
	var (
		dir   string
		files []string
	)

	if dir, err = config.Dir(); err != nil {
		return
	}

	if files, err = filepath.Glob(filepath.Join(dir, path, "*."+suffix)); err != nil {
		return
	}

	for _, f := range files {
		info := utilDir.FileInfo{}
		if info, err = utilDir.GetPathInfo(f); err != nil {
			return
		}

		cfg := &service.Config{}
		if err = config.ReadConfig(filepath.Join("services", info.BaseNoExt), info.ExtNoSpot, cfg); err != nil {
			return
		}

		if err = LoadService(ctx, cfg); err != nil {
			return
		}
	}

	return
}

func LoadService(ctx context.Context, config *service.Config, opts ...service.Option) (err error) {
	loadBalancer, err := lf.NewLoadBalancer(config.LoadBalancerStrategy)
	if err != nil {
		return
	}

	discovererOpts := []discoverer.OptionFunc{}
	if len(config.IDCNodes) > 0 {
		discovererOpts = append(discovererOpts, discoverer.WithIDCNodes(config.IDCNodes))
	}

	discovery, err := df.NewDiscoverer(config.DiscovererStrategy, config.ServiceName, loadBalancer, discovererOpts...)
	if err != nil {
		return
	}

	s, err := service.NewService(config, discovery, loadBalancer, opts...)
	if err != nil {
		return
	}
	if err = s.Start(ctx); err != nil {
		return err
	}

	if err = servicer.SetServicer(s); err != nil {
		return
	}

	return
}
