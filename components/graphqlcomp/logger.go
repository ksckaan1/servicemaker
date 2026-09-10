package graphqlcomp

import (
	"cmp"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/ksckaan1/logger"
)

func (f *GraphQL) restLoggerMW(ctx fiber.Ctx) error {
	start := time.Now()

	chainErr := ctx.Next()
	if chainErr != nil {
		if err := ctx.App().ErrorHandler(ctx, chainErr); err != nil {
			_ = ctx.SendStatus(fiber.StatusInternalServerError)
		}
	}

	if ctx.GetRespHeader("Content-Type") == "text/event-stream" /* SSE */ ||
		ctx.Response().StatusCode() == 101 /* WebSocket */ {
		return chainErr
	}

	latency := time.Since(start)
	ip := cmp.Or(ctx.Get("X-Client-Ip"), ctx.Get("Cf-Connecting-Ip"))
	statusCode := ctx.Response().StatusCode()

	messages := make([]any, 0, 10)

	messages = append(
		messages,
		"latency", latency.String(),
		"status_code", statusCode,
		"method", ctx.Method(),
		"url", ctx.OriginalURL(),
		"ip", ip,
		"x-forwarded-for", ctx.IPs(),
		"host", ctx.Hostname(),
		"request_body_length", len(ctx.Request().Body()),
		"response_body_length", len(ctx.Response().Body()),
		"user_agent", ctx.Get("User-Agent"),
		"content_type", ctx.Get("Content-Type"),
		"api", "graphql",
		"component", "graphql",
	)
	if chainErr != nil {
		messages = append(messages, "error", chainErr.Error())
	}

	switch {
	case statusCode >= 400:
		logger.Default.Error(ctx.Context(), "request received", messages...)
	default:
		logger.Default.Info(ctx.Context(), "request received", messages...)
	}

	return chainErr
}
