package main

import (
	"context"
	envoy_config_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_config_route_v3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	routeservice "github.com/envoyproxy/go-control-plane/envoy/service/route/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"time"
)

func main() {
	gopts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	addr := "localhost:19000"
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	conn, err := grpc.DialContext(ctx, addr, gopts...)
	if err != nil {
		log.Fatalln(err)
	}

	// 创建CDS grpc服务客户端
	client := routeservice.NewRouteDiscoveryServiceClient(conn)
	req := &discovery.DiscoveryRequest{
		Node: &envoy_config_core_v3.Node{
			Id: "test1",
		},
	}
	rsp, err := client.FetchRoutes(context.Background(), req)
	if err != nil {
		log.Fatalln(err)
	}

	// 获取resources节点内容
	log.Println("length=", len(rsp.GetResources()))
	for _, getResource := range rsp.GetResources() {
		// 反序列化为route对象
		routeConfig := &envoy_config_route_v3.RouteConfiguration{}
		err = getResource.UnmarshalTo(routeConfig)
		if err != nil {
			log.Fatalln(err)
		}
		log.Println(routeConfig)
	}
}
