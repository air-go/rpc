package orm

import (
	"fmt"
	"net/url"

	"github.com/pkg/errors"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/plugin/dbresolver"
)

type driverType string

const (
	driverMySQL    driverType = "mysql"
	driverPostgres driverType = "postgres"
)

const (
	defaultSSLMode  = "disable"
	defaultTimeZone = "Asia/Shanghai"
)

type Config struct {
	ServiceName string
	// Driver 为空时默认使用 mysql
	Driver driverType
	Master *instanceConfig
	Slave  *instanceConfig
}

type instanceConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DB       string
	Charset  string
	MaxOpen  int
	MaxIdle  int
	// SSLMode 仅 postgres 使用
	SSLMode string
	// TimeZone 对应 postgres 的 TimeZone、mysql 的 loc，为空时 mysql 不传 loc
	TimeZone string
}

type Orm struct {
	*gorm.DB
	config *gorm.Config
	tracer gorm.Plugin
}

type Option func(orm *Orm)

func WithTrace(tracer gorm.Plugin) Option {
	return func(orm *Orm) {
		orm.tracer = tracer
	}
}

func WithLogger(logger logger.Interface) Option {
	return func(orm *Orm) {
		orm.config.Logger = logger
	}
}

func NewOrm(cfg *Config, opts ...Option) (orm *Orm, err error) {
	orm = &Orm{
		config: &gorm.Config{
			SkipDefaultTransaction: true,
		},
	}

	for _, o := range opts {
		o(orm)
	}

	master := newDialector(cfg.Driver, cfg.Master)
	slave := newDialector(cfg.Driver, cfg.Slave)

	_orm, err := gorm.Open(master, orm.config)
	if err != nil {
		err = errors.Wrap(err, "open mysql conn error：")
		return nil, err
	}

	err = _orm.Use(orm.tracer)
	if err != nil {
		return
	}

	err = _orm.Use(dbresolver.Register(dbresolver.Config{
		Sources:  []gorm.Dialector{master},
		Replicas: []gorm.Dialector{slave},
		Policy:   dbresolver.RandomPolicy{},
	}).SetMaxOpenConns(cfg.Master.MaxOpen).SetMaxIdleConns(cfg.Master.MaxIdle))
	if err != nil {
		return
	}

	orm.DB = _orm

	return
}

func (orm *Orm) UseWrite() *gorm.DB {
	return orm.Clauses(dbresolver.Write)
}

func (orm *Orm) UseRead() *gorm.DB {
	return orm.Clauses(dbresolver.Read)
}

func newDialector(driver driverType, cfg *instanceConfig) gorm.Dialector {
	switch driver {
	case driverPostgres:
		return postgres.Open(getPostgresDSN(cfg))
	default:
		return mysql.Open(getMySQLDSN(cfg))
	}
}

func getMySQLDSN(cfg *instanceConfig) string {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=true",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DB,
		cfg.Charset)

	// loc 含 "/" 等字符，必须转义
	if cfg.TimeZone != "" {
		dsn += "&loc=" + url.QueryEscape(cfg.TimeZone)
	}

	return dsn
}

func getPostgresDSN(cfg *instanceConfig) string {
	sslMode := cfg.SSLMode
	if sslMode == "" {
		sslMode = defaultSSLMode
	}

	timeZone := cfg.TimeZone
	if timeZone == "" {
		timeZone = defaultTimeZone
	}

	return fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s sslmode=%s TimeZone=%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DB,
		sslMode,
		timeZone)
}
