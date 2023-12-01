package logger

import (
	"net/http"
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
	AppName        = "app_name"
	LogID          = "log_id"
	TraceID        = "trace_id"
	Module         = "module"
	ServiceName    = "service_name"
	RequestHeader  = "request_header"
	ResponseHeader = "response_header"
	Method         = "method"
	API            = "api"
	URI            = "uri"
	Request        = "request"
	Response       = "response"
	Code           = "code"
	ClientIP       = "client_ip"
	ClientPort     = "client_port"
	ServerIP       = "server_ip"
	ServerPort     = "server_port"
	Cost           = "cost"
	Errno          = "errno"
)

type Fields struct {
	AppName        string      `json:"app_name"`
	LogID          string      `json:"log_id"`
	TraceID        string      `json:"trace_id"`
	Module         string      `json:"module"`
	ServiceName    string      `json:"service_name"`
	RequestHeader  http.Header `json:"request_header"`
	ResponseHeader http.Header `json:"response_header"`
	Method         string      `json:"method"`
	API            string      `json:"api"`
	URI            string      `json:"uri"`
	Request        interface{} `json:"request"`
	Response       interface{} `json:"response"`
	Code           int         `json:"code"`
	ClientIP       string      `json:"client_ip"`
	ClientPort     int         `json:"client_port"`
	ServerIP       string      `json:"server_ip"`
	ServerPort     int         `json:"server_port"`
	Cost           int64       `json:"cost"`
	Error          string      `json:"error"`
	Stack          string      `json:"stack"`
}

var metaFields = map[string]struct{}{
	"app_name": {},
	"log_id":   {},
	"trace_id": {},
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
	return &field{key: "error", value: err.Error()}
}
