package gorm

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	gormLogger "gorm.io/gorm/logger"
)

func TestNewGorm(t *testing.T) {
	convey.Convey("TestNewGorm", t, func() {
		convey.Convey("success", func() {
			l, err := NewGorm(&GormConfig{})
			assert.Nil(t, err)
			assert.NotNil(t, l)
		})
	})
}

func TestGormLogger_LogMode(t *testing.T) {
	convey.Convey("TestGormLogger_LogMode", t, func() {
		convey.Convey("success", func() {
			l, err := NewGorm(&GormConfig{})
			assert.Nil(t, err)
			assert.NotNil(t, l)

			assert.Equal(t, &GormLogger{
				Logger:                    l.Logger,
				SlowThreshold:             l.SlowThreshold,
				LogLevel:                  gormLogger.Info,
				SkipCallerLookup:          l.SkipCallerLookup,
				IgnoreRecordNotFoundError: l.IgnoreRecordNotFoundError,
			}, l.LogMode(gormLogger.Info))
		})
	})
}

func TestGormLoggerWrite(t *testing.T) {
	convey.Convey("TestGormLoggerWrite", t, func() {
		convey.Convey("Info", func() {
			l, err := NewGorm(&GormConfig{})
			assert.Nil(t, err)
			assert.NotNil(t, l)

			l.Info(context.TODO(), "msg")
		})
		convey.Convey("Warn", func() {
			l, err := NewGorm(&GormConfig{})
			assert.Nil(t, err)
			assert.NotNil(t, l)

			l.Warn(context.TODO(), "msg")
		})
		convey.Convey("Error", func() {
			l, err := NewGorm(&GormConfig{})
			assert.Nil(t, err)
			assert.NotNil(t, l)

			l.Error(context.TODO(), "msg")
		})
	})
}

func TestGormLogger_Trace(t *testing.T) {
	fc := func() (string, int64) { return "select * from table;", 0 }
	convey.Convey("TestGormLogger_Trace", t, func() {
		convey.Convey("l.LogLevel <= gormLogger.Silent", func() {
			l, err := NewGorm(&GormConfig{})
			assert.Nil(t, err)
			assert.NotNil(t, l)

			l.Trace(context.TODO(), time.Now(), fc, nil)
		})
		convey.Convey("error log", func() {
			l, err := NewGorm(&GormConfig{
				Level:                     2,
				IgnoreRecordNotFoundError: true,
			})
			assert.Nil(t, err)
			assert.NotNil(t, l)

			l.Trace(context.TODO(), time.Now(), fc, errors.New("error"))
		})
		convey.Convey("slow log", func() {
			l, err := NewGorm(&GormConfig{
				Level:                     2,
				IgnoreRecordNotFoundError: true,
				SlowThreshold:             1,
			})
			assert.Nil(t, err)
			assert.NotNil(t, l)

			l.Trace(context.TODO(), time.Now().Add(-time.Second), fc, nil)
		})
		convey.Convey("info log", func() {
			l, err := NewGorm(&GormConfig{
				Level:                     2,
				IgnoreRecordNotFoundError: true,
				SlowThreshold:             0,
			})
			assert.Nil(t, err)
			assert.NotNil(t, l)

			l.Trace(context.TODO(), time.Now(), fc, nil)
		})
	})
}

func TestGormLogger_logger(t *testing.T) {
	convey.Convey("TestGormLogger_logger", t, func() {
		convey.Convey("caller file contains gorm.io", func() {
			l, err := NewGorm(&GormConfig{})
			assert.Nil(t, err)
			assert.NotNil(t, l)

			zl := l.logger()
			assert.NotEmpty(t, zl)
		})
	})
}
