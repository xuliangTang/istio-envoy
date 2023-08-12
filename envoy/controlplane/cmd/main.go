package main

import (
	"context"
	"fmt"
	clusterservice "github.com/envoyproxy/go-control-plane/envoy/service/cluster/v3"
	listenerservice "github.com/envoyproxy/go-control-plane/envoy/service/listener/v3"
	routeservice "github.com/envoyproxy/go-control-plane/envoy/service/route/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/envoyproxy/go-control-plane/pkg/server/v3"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"istio-envoy/envoy/controlplane/utils"
	"log"
	"net"
	"os"
	"time"
)

func main() {
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

	// envoy 配置的缓存快照
	snapshot := utils.GenerateSnapshot(utils.SnapshotVersion)
	if err := snapshot.Consistent(); err != nil {
		llog.Errorf("snapshot inconsistency: %+v\n%+v", snapshot, err)
		os.Exit(1)
	}

	// Add the snapshot to the cache
	// nodeID 必须要设置
	nodeID := "test1"
	if err := snapshotCache.SetSnapshot(context.Background(), nodeID, *snapshot); err != nil {
		os.Exit(1)
	}

	// 请求回调
	cb := &utils.Callbacks{Debug: llog.Debug}
	// 官方提供的控制面server
	srv := server.NewServer(context.Background(), snapshotCache, cb)
	// 注册集群服务CDS LDS RDS
	clusterservice.RegisterClusterDiscoveryServiceServer(grpcServer, srv)
	listenerservice.RegisterListenerDiscoveryServiceServer(grpcServer, srv)
	routeservice.RegisterRouteDiscoveryServiceServer(grpcServer, srv)

	errCh := make(chan error)
	// 启动grpc服务
	go func() {
		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", 19000))
		if err != nil {
			errCh <- err
			return
		}
		if err = grpcServer.Serve(lis); err != nil {
			errCh <- err
			return
		}
	}()

	// 启动http服务
	go func() {
		r := gin.New()

		r.GET("/update", func(c *gin.Context) {
			// 模拟提交yaml 修改envoy配置
			utils.UpstreamHost = "172.17.0.3"
			utils.SnapshotVersion = "v2-" + time.Now().Format("2006-01-02 15:04:05")
			snapshot := utils.GenerateSnapshot(utils.SnapshotVersion)
			if err := snapshotCache.SetSnapshot(c, nodeID, *snapshot); err != nil {
				c.String(400, err.Error())
				return
			}

			c.String(200, "success")
		})

		if err := r.Run(":18000"); err != nil {
			errCh <- err
		}
	}()

	err := <-errCh
	log.Println(err)
	os.Exit(1)
}
