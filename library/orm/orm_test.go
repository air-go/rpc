package orm

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"

	logger "github.com/air-go/rpc/library/logger/zap/gorm"
	"github.com/air-go/rpc/library/opentracing/gorm"
)

func TestNewOrm(t *testing.T) {
	convey.Convey("TestNewOrm", t, func() {
		convey.Convey("success", func() {
			l, _ := logger.NewGorm(&logger.GormConfig{})
			g, err := NewOrm(&Config{Master: &instanceConfig{}, Slave: &instanceConfig{}},
				WithTrace(gorm.GormTrace),
				WithLogger(l))

			assert.Nil(t, g)
			assert.NotNil(t, err)
		})

		convey.Convey("postgres", func() {
			l, _ := logger.NewGorm(&logger.GormConfig{})
			g, err := NewOrm(&Config{Driver: driverPostgres, Master: &instanceConfig{}, Slave: &instanceConfig{}},
				WithTrace(gorm.GormTrace),
				WithLogger(l))

			assert.Nil(t, g)
			assert.NotNil(t, err)
		})
	})
}

func TestGetDSN(t *testing.T) {
	convey.Convey("TestGetDSN", t, func() {
		cfg := &instanceConfig{
			Host:     "127.0.0.1",
			Port:     "3306",
			User:     "root",
			Password: "123456",
			DB:       "test",
			Charset:  "utf8mb4",
		}

		convey.Convey("mysql", func() {
			assert.Equal(t, "root:123456@tcp(127.0.0.1:3306)/test?charset=utf8mb4&parseTime=true",
				getMySQLDSN(cfg))
		})

		convey.Convey("mysql with loc", func() {
			c := *cfg
			c.TimeZone = "Asia/Shanghai"
			assert.Equal(t, "root:123456@tcp(127.0.0.1:3306)/test?charset=utf8mb4&parseTime=true&loc=Asia%2FShanghai",
				getMySQLDSN(&c))
		})

		convey.Convey("postgres default", func() {
			assert.Equal(t, "user=root password=123456 host=127.0.0.1 port=3306 dbname=test sslmode=disable TimeZone=Asia/Shanghai",
				getPostgresDSN(cfg))
		})

		convey.Convey("postgres custom", func() {
			c := *cfg
			c.SSLMode = "require"
			c.TimeZone = "UTC"
			assert.Equal(t, "user=root password=123456 host=127.0.0.1 port=3306 dbname=test sslmode=require TimeZone=UTC",
				getPostgresDSN(&c))
		})
	})
}

func TestNewDialector(t *testing.T) {
	convey.Convey("TestNewDialector", t, func() {
		cfg := &instanceConfig{}

		convey.Convey("default is mysql", func() {
			assert.Equal(t, "mysql", newDialector("", cfg).Name())
			assert.Equal(t, "mysql", newDialector(driverMySQL, cfg).Name())
			assert.Equal(t, "mysql", newDialector("unknown", cfg).Name())
		})

		convey.Convey("postgres", func() {
			assert.Equal(t, "postgres", newDialector(driverPostgres, cfg).Name())
		})
	})
}
