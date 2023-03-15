//go:generate go run -mod=mod github.com/golang/mock/mockgen -package leakybucket -destination=./leaky_bucket_mock.go  -source=leaky_bucket.go -build_flags=-mod=mod
package leakybucket
