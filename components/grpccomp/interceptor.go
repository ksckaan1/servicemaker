package grpccomp

import (
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

// serverOptions builds the server options every server instance needs. The
// interceptors are the only part that changes between rebuilds, so the base
// options live here rather than being repeated in each builder.
func (g *GRPC) serverOptions() []grpc.ServerOption {
	return []grpc.ServerOption{
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(g.unaryServerInterceptors...),
		grpc.ChainStreamInterceptor(g.streamServerInterceptors...),
	}
}

func (g *GRPC) AddUnaryServerInterceptor(interceptor grpc.UnaryServerInterceptor) {
	g.unaryServerInterceptors = append(g.unaryServerInterceptors, interceptor)
	g.server = grpc.NewServer(g.serverOptions()...)
}

func (g *GRPC) AddStreamServerInterceptor(interceptor grpc.StreamServerInterceptor) {
	g.streamServerInterceptors = append(g.streamServerInterceptors, interceptor)
	g.server = grpc.NewServer(g.serverOptions()...)
}
