//go:generate go run -mod=mod github.com/golang/mock/mockgen -package limiter -destination=./limiter_mock.go  -source=limiter.go -build_flags=-mod=mod
package limiter
