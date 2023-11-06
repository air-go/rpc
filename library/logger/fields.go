package logger

import (
	"net/http"
	"time"
)

const (
	LogHeader = "Log-Id"
)

const (
	ModuleHTTP  = "HTTP"
	ModuleRPC   = "RPC"
	ModuleMySQL = "MySQL"
	ModuleRedis = "Redis"
	ModuleQueue = "Queue"
	ModuleCron  = "Cron"
)

const (
	AppName     = "app_name"
	Module      = "module"
	ServiceName = "service_name"
	LogID       = "log_id"
	TraceID     = "trace_id"
	Header      = "header"
	Method      = "method"
	Request     = "request"
	Response    = "response"
	Code        = "code"
	ClientIP    = "client_ip"
	ClientPort  = "client_port"
	ServerIP    = "server_ip"
	ServerPort  = "server_port"
	API         = "api"
	URI         = "uri"
	Cost        = "cost"
	Timeout     = "timeout"
	Trace       = "trace"
)

type Fields struct {
	AppName     string        `json:"app_name"`
	Module      string        `json:"module"`
	ServiceName string        `json:"service_name"`
	LogID       string        `json:"log_id"`
	TraceID     string        `json:"trace_id"`
	Header      http.Header   `json:"header"`
	Method      string        `json:"method"`
	Request     interface{}   `json:"request"`
	Response    interface{}   `json:"response"`
	Code        int           `json:"code"`
	ClientIP    string        `json:"client_ip"`
	ClientPort  int           `json:"client_port"`
	ServerIP    string        `json:"server_ip"`
	ServerPort  int           `json:"server_port"`
	API         string        `json:"api"`
	URI         string        `json:"uri"`
	Cost        int64         `json:"cost"`
	Timeout     time.Duration `json:"timeout"`
	Trace       string        `json:"trace"`
}

type Field interface {
	Key() string
	Value() any
}

type field struct {
	key   string
	value any
}

func (f *field) Key() string {
	return f.key
}

func (f *field) Value() any {
	return f.value
}

func Reflect(key string, value any) Field {
	return &field{key: key, value: value}
}

func Error(err error) Field {
	return &field{key: "error", value: err}
}
