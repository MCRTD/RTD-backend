package middleware

import "github.com/danielgtaylor/huma/v2"

func Corsfunc(ctx huma.Context, next func(huma.Context)) {
	ctx.AppendHeader("Access-Control-Allow-Origin", "http://localhost:5173")
	ctx.AppendHeader("Access-Control-Allow-Credentials", "true")
	ctx.AppendHeader("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	ctx.AppendHeader("Access-Control-Allow-Headers", "Content-Type, Authorization")
	next(ctx)
}
