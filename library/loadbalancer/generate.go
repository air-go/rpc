//go:generate go run -mod=mod github.com/golang/mock/mockgen -package loadbalancer -destination=./load_balancer_mock.go  -source=load_balancer.go -build_flags=-mod=mod
package loadbalancer
