package otel

import "go.opentelemetry.io/otel/attribute"

const (
	Baggage     string = "baggage"
	LibraryName string = "rpc.otel"
)

const (
	AttributeTraceID = attribute.Key("trace.id")
)

const (
	AttributeSpanID = attribute.Key("span.id")
)

const (
	AttributeLogID    = attribute.Key("log_id")
	AttributeRequest  = attribute.Key("request")
	AttributeResponse = attribute.Key("response")
)

const (
	AttributeGinError = attribute.Key("gin.errors")
)

const (
	AttributeRedisError     = attribute.Key("redis.cmd.error")
	AttributeRedisCmdName   = attribute.Key("redis.cmd.name")
	AttributeRedisCmdString = attribute.Key("redis.cmd.string")
	AttributeRedisCmdArgs   = attribute.Key("redis.cmd.args")
)

const (
	TracerNameHTTPServer = "http_server"
	TracerNameHTTPClient = "http_client"
	TracerNameGorm       = "grom"
	TracerNameRedis      = "redis"
)
