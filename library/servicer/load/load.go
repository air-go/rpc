package load

import (
	"path/filepath"

	utilDir "github.com/why444216978/go-util/dir"

	"github.com/air-go/rpc/library/config"
	df "github.com/air-go/rpc/library/discoverer/factory"
	"github.com/air-go/rpc/library/etcd"
	lf "github.com/air-go/rpc/library/loadbalancer/factory"
	"github.com/air-go/rpc/library/servicer"
	"github.com/air-go/rpc/library/servicer/service"
)

func LoadGlobPattern(path, suffix string, etcd *etcd.Etcd) (err error) {
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

	info := utilDir.FileInfo{}
	cfg := &service.Config{}
	for _, f := range files {
		if info, err = utilDir.GetPathInfo(f); err != nil {
			return
		}
		if err = config.ReadConfig(filepath.Join("services", info.BaseNoExt), info.ExtNoSpot, cfg); err != nil {
			return
		}

		if err = LoadService(cfg); err != nil {
			return
		}
	}

	return
}

func LoadService(config *service.Config, opts ...service.Option) (err error) {
	loadBalancer, err := lf.NewLoadBalancer(config.LoadBalancerStrategy)
	if err != nil {
		return
	}

	discovery, err := df.NewDiscoverer(config.DiscovererStrategy, config.ServiceName, loadBalancer)
	if err != nil {
		return
	}

	s, err := service.NewService(config, discovery, loadBalancer, opts...)
	if err != nil {
		return
	}

	if err = servicer.SetServicer(s); err != nil {
		return
	}

	return
}
