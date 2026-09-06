package grpccomp

import (
	"context"
	"time"

	"github.com/ksckaan1/logger"
	"google.golang.org/grpc"
)

func (g *GRPC) loggerMWUnary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		startTime := time.Now()

		resp, err := handler(ctx, req)
		if err == nil {
			logger.Default.Info(
				ctx, "request received",
				"method", info.FullMethod,
				"took", time.Since(startTime).String(),
				"api", "grpc",
			)
		} else {
			logger.Default.Error(
				ctx, "request received",
				"method", info.FullMethod,
				"took", time.Since(startTime).String(),
				"api", "grpc",
				"error", err,
			)
		}

		return resp, err
	}
}

func (g *GRPC) loggerMWStream() grpc.StreamServerInterceptor {
	return func(srv any, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		startTime := time.Now()

		err := handler(srv, stream)
		if err == nil {
			logger.Default.Info(
				stream.Context(), "grpc api stream",
				"method", info.FullMethod,
				"took", time.Since(startTime).String(),
				"api", "grpc",
			)
		} else {
			logger.Default.Error(
				stream.Context(), "grpc api stream failed",
				"method", info.FullMethod,
				"took", time.Since(startTime).String(),
				"error", err,
				"api", "grpc",
			)
		}

		return err
	}
}
