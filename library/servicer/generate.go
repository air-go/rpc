//go:generate go run -mod=mod github.com/golang/mock/mockgen -package servicer -destination=./servicer_mock.go  -source=servicer.go -build_flags=-mod=mod
package servicer
