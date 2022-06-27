package main

import (
	"context"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	gw "learningGRPC/my_guest_reverse_proxy/grpc/cdp/v1/my_guest"
	"net/http"
)

func main() {
	// 初始化ctx
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// 初始化newServeMux
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	// 在端点注册Guest Service Handler
	err := gw.RegisterGuestServicesHandlerFromEndpoint(ctx, mux, "localhost:54321", opts)
	if err != nil {
		panic(err)
	}

	// 监听http服务s
	if err := http.ListenAndServe(":54322", mux); err != nil {
		panic(err)
	}
}
