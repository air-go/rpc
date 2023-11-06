package app

import (
	"time"

	"github.com/why444216978/go-util/sys"

	"github.com/air-go/rpc/library/config"
)

var app struct {
	AppName        string
	RegistryName   string
	LocalIP        string
	AppPort        int
	Pprof          bool
	IsDebug        bool
	ContextTimeout int
	ConnectTimeout int
	WriteTimeout   int
	ReadTimeout    int
}

func InitApp() (err error) {
	err = config.ReadConfig("app", "toml", &app)
	app.LocalIP, _ = sys.LocalIP()
	return
}

func Name() string {
	return app.AppName
}

func RegistryName() string {
	return app.RegistryName
}

func LocalIP() string {
	return app.LocalIP
}

func Port() int {
	return app.AppPort
}

func Pprof() bool {
	return app.Pprof
}

func Debug() bool {
	return app.IsDebug
}

func ContextTimeout() time.Duration {
	if app.ContextTimeout == 0 {
		return time.Duration(1000) * time.Millisecond
	}
	return time.Duration(app.ContextTimeout) * time.Millisecond
}

func ConnectTimeout() time.Duration {
	if app.ConnectTimeout == 0 {
		return time.Duration(1000) * time.Millisecond
	}
	return time.Duration(app.ConnectTimeout) * time.Millisecond
}

func WriteTimeout() time.Duration {
	if app.WriteTimeout == 0 {
		return time.Duration(1000) * time.Millisecond
	}
	return time.Duration(app.WriteTimeout) * time.Millisecond
}

func ReadTimeout() time.Duration {
	if app.ReadTimeout == 0 {
		return time.Duration(1000) * time.Millisecond
	}
	return time.Duration(app.ReadTimeout) * time.Millisecond
}
