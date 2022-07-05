package agollo

import (
	"context"
	"io/fs"
	"os"
	"strings"

	"github.com/apolloconfig/agollo/v4"
	"github.com/apolloconfig/agollo/v4/agcache"
	"github.com/apolloconfig/agollo/v4/cluster"
	"github.com/apolloconfig/agollo/v4/component/log"
	"github.com/apolloconfig/agollo/v4/env/config"
	"github.com/apolloconfig/agollo/v4/env/file"
	"github.com/apolloconfig/agollo/v4/protocol/auth"

	"github.com/air-go/rpc/library/apollo/listener"
)

const (
	defaultBackupConfigFileMode fs.FileMode = 0744
	defaultIsBackupConfig                   = true
	defaultBackupConfigPath                 = ".apollo_config"
)

type Option func(*ApolloClient)

func WithSecret(secret string) Option {
	return func(a *ApolloClient) { a.appConfig.Secret = secret }
}

func WithIsBackupConfig(isBackup bool) Option {
	return func(a *ApolloClient) { a.appConfig.IsBackupConfig = isBackup }
}

func WithBackupConfigPath(backupPath string) Option {
	return func(a *ApolloClient) { a.appConfig.BackupConfigPath = backupPath }
}

func WithSyncServerTimeout(syncServerTimeout int) Option {
	return func(a *ApolloClient) { a.appConfig.SyncServerTimeout = syncServerTimeout }
}

func WithCustomListeners(listeners []listener.Listener) Option {
	return func(a *ApolloClient) { a.listeners = listeners }
}

type ApolloClient struct {
	client         agollo.Client
	listeners      []listener.Listener
	appConfig      *config.AppConfig
	backupFileMode fs.FileMode
}

var defaultAppConfig = func(appID, ip, cluster string, namespaces []string) *config.AppConfig {
	return &config.AppConfig{
		AppID:             appID,
		NamespaceName:     strings.Join(namespaces, ","),
		IP:                ip,
		Cluster:           cluster,
		IsBackupConfig:    defaultIsBackupConfig,
		BackupConfigPath:  defaultBackupConfigPath,
		SyncServerTimeout: 10,
	}
}

func New(ctx context.Context, appID, ip, cluster string, namespaces []string, opts ...Option) (err error) {
	ac := &ApolloClient{
		appConfig: defaultAppConfig(appID, ip, cluster, namespaces),
	}

	for _, o := range opts {
		o(ac)
	}

	if ac.appConfig.AppID == "" {
		panic("appID empty")
	}

	if ac.appConfig.IP == "" {
		panic("ip empty")
	}

	if ac.appConfig.Cluster == "" {
		panic("cluster empty")
	}

	if ac.backupFileMode == 0 {
		ac.backupFileMode = defaultBackupConfigFileMode
	}

	// init back file
	err = os.MkdirAll(ac.appConfig.BackupConfigPath, ac.backupFileMode)
	if err != nil {
		return
	}

	// init client
	client, err := agollo.StartWithConfig(func() (*config.AppConfig, error) { return ac.appConfig, nil })
	if err != nil {
		return
	}

	// init listener
	for _, l := range ac.listeners {
		client.AddChangeListener(l)
		l.InitConfig(client)
	}

	return
}

// SetSignature set custom http auth
func SetSignature(a auth.HTTPAuth) {
	agollo.SetSignature(a)
}

// SetBackupFileHandler set custom backup file handler
func SetBackupFileHandler(f file.FileHandler) {
	agollo.SetBackupFileHandler(f)
}

// SetLoadBalance set custom load balance
func SetLoadBalance(l cluster.LoadBalance) {
	agollo.SetLoadBalance(l)
}

// SetLogger set custom logger
func SetLogger(l log.LoggerInterface) {
	agollo.SetLogger(l)
}

// SetCache set custom cache factory
func SetCacheFactory(f agcache.CacheFactory) {
	agollo.SetCache(f)
}
