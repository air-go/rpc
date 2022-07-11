//go:generate go run -mod=mod github.com/golang/mock/mockgen -package redis -source /home/users/weihaoyu/go/pkg/mod/github.com/go-redis/redis/v8@v8.11.4/commands.go -destination ./mock/redis/redis.go Cmdable
package mock
