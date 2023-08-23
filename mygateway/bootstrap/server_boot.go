package bootstrap

import (
	"context"
	"fmt"
	routeservice "github.com/envoyproxy/go-control-plane/envoy/service/route/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/envoyproxy/go-control-plane/pkg/server/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"istio-envoy/mygateway/utils"
	"net"
	"os"
	"time"
)

type GatewayBooter struct{}

func NewGatewayBooter() *GatewayBooter {
	return &GatewayBooter{}
}

func (*GatewayBooter) Start(context.Context) error {
	return runXdsServer()
}

// 启动xds server
func runXdsServer() error {
	var grpcOptions []grpc.ServerOption
	grpcOptions = append(grpcOptions,
		grpc.MaxConcurrentStreams(1000), // 一条GRPC连接允许并发的发送和接收多个Stream
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:    time.Second * 30, // 连接超过多少时间不活跃，则会去探测是否依然alive
			Timeout: time.Second * 5,
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             time.Second * 30, // 发送ping之前最少要等待时间
			PermitWithoutStream: true,             // 连接空闲时仍然发送PING帧监测
		}),
	)

	// 创建grpc服务
	grpcServer := grpc.NewServer(grpcOptions...)

	// 日志
	llog := utils.MyLogger{}
	// 创建缓存系统
	snapshotCache := cache.NewSnapshotCache(false, cache.IDHash{}, llog)

	// envoy配置的缓存快照
	snapshot := utils.GenerateSnapshot("1")
	if err := snapshot.Consistent(); err != nil {
		llog.Errorf("snapshot inconsistency: %+v\n%+v", snapshot, err)
		os.Exit(1)
	}

	// Add the snapshot to the cache
	// nodeID 必须要设置
	nodeID := "test1"
	if err := snapshotCache.SetSnapshot(context.Background(), nodeID, snapshot); err != nil {
		os.Exit(1)
	}

	// 请求回调
	cb := &utils.Callbacks{Debug: llog.Debug}
	// 官方提供的控制面server
	srv := server.NewServer(context.Background(), snapshotCache, cb)
	// 注册
	routeservice.RegisterRouteDiscoveryServiceServer(grpcServer, srv)

	// 启动服务
	fmt.Println("启动xDS服务")
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", 19000))
	if err != nil {
		return err
	}
	if err = grpcServer.Serve(lis); err != nil {
		return err
	}

	return nil
}
