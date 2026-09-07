package grpccomp

import (
	"context"
	"fmt"
	"net"

	"github.com/ksckaan1/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type GRPC struct {
	server *grpc.Server

	// interceptors
	unaryServerInterceptors  []grpc.UnaryServerInterceptor
	streamServerInterceptors []grpc.StreamServerInterceptor

	// CONFIGS
	Addr    string `env:"GRPC_ADDR"`
	Reflect bool   `env:"GRPC_REFLECT" envDefault:"true"`
}

func (g *GRPC) Init(ctx context.Context) error {
	g.unaryServerInterceptors = []grpc.UnaryServerInterceptor{g.loggerMWUnary()}
	g.streamServerInterceptors = []grpc.StreamServerInterceptor{g.loggerMWStream()}

	opts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(g.unaryServerInterceptors...),
		grpc.ChainStreamInterceptor(g.streamServerInterceptors...),
	}

	g.server = grpc.NewServer(opts...)

	logger.Default.Info(ctx, "grpc server initialized")

	return nil
}

func (g *GRPC) Run(ctx context.Context) error {
	if g.Reflect {
		reflection.Register(g.server)
	}

	listener, err := net.Listen("tcp", g.Addr)
	if err != nil {
		return fmt.Errorf("net.Listen: %w", err)
	}

	logger.Default.Info(
		ctx, "grpc server listening",
		"addr", g.Addr,
	)

	go func() {
		<-ctx.Done()
		g.server.GracefulStop()
	}()

	err = g.server.Serve(listener)
	if err != nil {
		return fmt.Errorf("g.server.Serve: %w", err)
	}

	return nil
}

func (g *GRPC) Close(ctx context.Context) error {
	g.server.GracefulStop()
	logger.Default.Info(ctx, "grpc server stopped")
	return nil
}

func (g *GRPC) Server() *grpc.Server {
	return g.server
}
