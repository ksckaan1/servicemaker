package fibercomp

import (
	"context"
	"encoding/json/v2"
	"fmt"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/ksckaan1/logger"
)

type Fiber struct {
	router *fiber.App

	// CONFIGS
	BodyLimit                    int      `env:"FIBER_BODY_LIMIT" envDefault:"0"`
	MaxRanges                    int      `env:"FIBER_MAX_RANGES" envDefault:"0"`
	Concurrency                  int      `env:"FIBER_CONCURRENCY" envDefault:"0"`
	StreamRequestBody            bool     `env:"FIBER_STREAM_REQUEST_BODY" envDefault:"false"`
	DisablePreParseMultipartForm bool     `env:"FIBER_DISABLE_PRE_PARSE_MULTIPART_FORM" envDefault:"false"`
	Addr                         string   `env:"FIBER_ADDR" envDefault:":8080"`
	DisableStartupMessage        bool     `env:"FIBER_DISABLE_STARTUP_MESSAGE" envDefault:"true"`
	UseRecoverMW                 bool     `env:"FIBER_USE_RECOVER_MW" envDefault:"true"`
	CORSAllowedOrigins           []string `env:"FIBER_CORS_ALLOWED_ORIGINS"`
	CORSAllowedMethods           []string `env:"FIBER_CORS_ALLOWED_METHODS" envDefault:"GET,POST,PUT,DELETE,PATCH,OPTIONS"`
	CORSAllowedHeaders           []string `env:"FIBER_CORS_ALLOWED_HEADERS" envDefault:"Origin,Content-Type,Accept,Authorization,Cookie"`
	CORSExposeHeaders            []string `env:"FIBER_CORS_EXPOSE_HEADERS" envDefault:"Content-Length,Content-Type,Set-Cookie"`
	CORSMaxAge                   int      `env:"FIBER_CORS_MAX_AGE" envDefault:"86400"`
	CORSAllowCredentials         bool     `env:"FIBER_CORS_ALLOW_CREDENTIALS" envDefault:"false"`
	CORSDisableValueRedaction    bool     `env:"FIBER_CORS_DISABLE_VALUE_REDUCTION" envDefault:"false"`
	CORSAllowPrivateNetwork      bool     `env:"FIBER_CORS_ALLOW_PRIVATE_NETWORK" envDefault:"false"`
}

func (f *Fiber) Init(ctx context.Context) error {
	f.router = fiber.New(fiber.Config{
		BodyLimit:                    f.BodyLimit,
		MaxRanges:                    f.MaxRanges,
		Concurrency:                  f.Concurrency,
		StreamRequestBody:            f.StreamRequestBody,
		DisablePreParseMultipartForm: f.DisablePreParseMultipartForm,
		JSONEncoder: func(v any) ([]byte, error) {
			return json.Marshal(v)
		},
		JSONDecoder: func(data []byte, v any) error {
			return json.Unmarshal(data, v)
		},
	})

	if f.UseRecoverMW {
		f.router.Use(recover.New())
	}

	if len(f.CORSAllowedOrigins) > 0 {
		f.router.Use(cors.New(cors.Config{
			AllowOrigins:          f.CORSAllowedOrigins,
			AllowMethods:          f.CORSAllowedMethods,
			AllowHeaders:          f.CORSAllowedHeaders,
			ExposeHeaders:         f.CORSExposeHeaders,
			MaxAge:                f.CORSMaxAge,
			DisableValueRedaction: f.CORSDisableValueRedaction,
			AllowCredentials:      f.CORSAllowCredentials,
			AllowPrivateNetwork:   f.CORSAllowPrivateNetwork,
		}))
	}

	f.router.Use(f.restLoggerMW)

	logger.Default.Info(ctx, "fiber initialized")

	return nil
}

func (f *Fiber) Run(ctx context.Context) error {
	f.router.Hooks().OnListen(func(data fiber.ListenData) error {
		logger.Default.Info(
			ctx, "fiber listening",
			"host", data.Host,
			"port", data.Port,
		)
		return nil
	})
	err := f.router.Listen(f.Addr, fiber.ListenConfig{
		GracefulContext:       ctx,
		DisableStartupMessage: f.DisableStartupMessage,
	})
	if err != nil {
		return fmt.Errorf("router.Listen: %w", err)
	}

	return nil
}

func (f *Fiber) Close(ctx context.Context) error {
	err := f.router.Shutdown()
	if err != nil {
		return fmt.Errorf("router.Shutdown: %w", err)
	}
	logger.Default.Info(ctx, "fiber closed")
	return nil
}

func (f *Fiber) Router() *fiber.App {
	return f.router
}
