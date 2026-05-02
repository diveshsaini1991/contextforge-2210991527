package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	cfmcp "github.com/diveshsaini1991/contextforge-2210991527/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	mcpServer := cfmcp.NewContextForgeServer()
	errLog := log.New(os.Stderr, "[contextforge] ", log.LstdFlags)

	if err := server.ServeStdio(mcpServer,
		server.WithErrorLogger(errLog),
		server.WithStdioContextFunc(func(_ context.Context) context.Context {
			return ctx
		}),
	); err != nil {
		errLog.Fatalf("server error: %v", err)
	}
}
