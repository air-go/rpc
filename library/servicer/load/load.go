package load

import (
	"errors"
	"path/filepath"
	"strings"

	"github.com/why444216978/go-util/assert"
	utilDir "github.com/why444216978/go-util/dir"

	"github.com/air-go/rpc/library/config"
	"github.com/air-go/rpc/library/etcd"
	"github.com/air-go/rpc/library/registry"
	registryEtcd "github.com/air-go/rpc/library/registry/etcd"
	"github.com/air-go/rpc/library/selector/factory"
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

	var discover registry.Discovery
	info := utilDir.FileInfo{}
	cfg := &service.Config{}
	for _, f := range files {
		if info, err = utilDir.GetPathInfo(f); err != nil {
			return
		}
		if err = config.ReadConfig(filepath.Join("services", info.BaseNoExt), info.ExtNoSpot, cfg); err != nil {
			return
		}

		if cfg.Type == servicer.TypeRegistry {
			if assert.IsNil(etcd) {
				return errors.New("LoadGlobPattern etcd nil")
			}
			if strings.TrimSpace(cfg.RegistryName) == "" {
				return errors.New("service RegistryName is empty")
			}

			if discover, err = registryEtcd.NewDiscovery(etcd.Client, cfg.RegistryName); err != nil {
				return
			}
		}

		if err = LoadService(cfg, service.WithDiscovery(discover), service.WithSelector(factory.New(cfg.ServiceName, cfg.Selector))); err != nil {
			return
		}
	}

	return
}

func LoadService(config *service.Config, opts ...service.Option) (err error) {
	s, err := service.NewService(config, opts...)
	if err != nil {
		return
	}

	servicer.SetServicer(s)

	return nil
}
