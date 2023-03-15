//go:generate go run -mod=mod github.com/golang/mock/mockgen -package slidinglog -destination=./sliding_log_mock.go  -source=sliding_log.go -build_flags=-mod=mod
package slidinglog
