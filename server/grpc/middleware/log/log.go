package log

import (
	"context"
	"net"
	"time"

	"github.com/why444216978/go-util/snowflake"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"

	"github.com/air-go/rpc/library/logger"
)

func LogIDFromMD(md metadata.MD) string {
	logIDs := md.Get(logger.LogID)
	if len(logIDs) > 0 && logIDs[0] != "" {
		return logIDs[0]
	}

	return snowflake.Generate().String()
}

func GetPeerAddr(ctx context.Context) string {
	var addr string
	if pr, ok := peer.FromContext(ctx); ok {
		if tcpAddr, ok := pr.Addr.(*net.TCPAddr); ok {
			addr = tcpAddr.IP.String()
		} else {
			addr = pr.Addr.String()
		}
	}
	return addr
}

func UnaryServerInterceptor(l logger.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		ctx = logger.InitFieldsContainer(ctx)

		start := time.Now()
		resp, err = handler(ctx, req)
		if l == nil {
			return
		}

		md, has := metadata.FromIncomingContext(ctx)
		if !has {
			md = metadata.MD{}
		}
		logID := LogIDFromMD(md)

		grpc.SetTrailer(ctx, metadata.MD{
			logger.LogID: []string{logID},
		})

		// TODO full fields
		logger.AddField(ctx,
			logger.Reflect(logger.LogID, logID),
			logger.Reflect(logger.Cost, time.Since(start).Milliseconds()))
		if err != nil {
			l.Error(ctx, "grpc err", logger.Error(err))
		} else {
			l.Info(ctx, "grpc info")
		}

		return
	}
}

func UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
		// TODO
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}
