package main

import (
	"context"
	"fmt"
	envoy_config_cluster_v3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_config_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	clusterservice "github.com/envoyproxy/go-control-plane/envoy/service/cluster/v3"
	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"time"
)

// 模拟envoy请求控制面
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
	client := clusterservice.NewClusterDiscoveryServiceClient(conn)
	req := &discovery.DiscoveryRequest{
		Node: &envoy_config_core_v3.Node{
			Id: "test1",
		},
	}
	rsp, err := client.FetchClusters(context.Background(), req)
	if err != nil {
		log.Fatalln(err)
	}

	// 获取resources节点内容
	getResource := rsp.GetResources()[0]
	// 反序列化为cluster对象
	cluster := &envoy_config_cluster_v3.Cluster{}
	err = getResource.UnmarshalTo(cluster)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(cluster)
}
