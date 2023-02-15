//go:generate go run -mod=mod github.com/golang/mock/mockgen -package redis -destination ./mock/third/redis/redis.go  github.com/go-redis/redis/v8 Cmdable
package generate
