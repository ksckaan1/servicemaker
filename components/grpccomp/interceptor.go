package grpccomp

import "google.golang.org/grpc"

func (g *GRPC) AddUnaryServerInterceptor(interceptor grpc.UnaryServerInterceptor) {
	g.unaryServerInterceptors = append(g.unaryServerInterceptors, interceptor)

	opts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(g.unaryServerInterceptors...),
		grpc.ChainStreamInterceptor(g.streamServerInterceptors...),
	}

	g.server = grpc.NewServer(opts...)
}

func (g *GRPC) AddStreamServerInterceptor(interceptor grpc.StreamServerInterceptor) {
	g.streamServerInterceptors = append(g.streamServerInterceptors, interceptor)

	opts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(g.unaryServerInterceptors...),
		grpc.ChainStreamInterceptor(g.streamServerInterceptors...),
	}

	g.server = grpc.NewServer(opts...)
}
