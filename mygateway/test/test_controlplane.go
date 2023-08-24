package main

import (
	"context"
	"istio-envoy/mygateway/bootstrap"
)

func main() {
	// 单独启动控制面程序
	boot := bootstrap.NewGatewayBooter()
	boot.Start(context.Background())
}
