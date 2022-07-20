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
	})
}
