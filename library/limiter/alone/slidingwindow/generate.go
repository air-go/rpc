//go:generate go run -mod=mod github.com/golang/mock/mockgen -package slidingwindow -destination=./sliding_window_mock.go  -source=sliding_window.go -build_flags=-mod=mod
package slidingwindow
