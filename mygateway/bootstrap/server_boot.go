package bootstrap

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
	"istio-envoy/mygateway/utils"
	v1 "k8s.io/api/networking/v1"
	"log"
	"net"
	"os"
	"strconv"
	"time"
)

type GatewayBooter struct{}

func NewGatewayBooter() *GatewayBooter {
	return &GatewayBooter{}
}

func (*GatewayBooter) Start(context.Context) error {
	utils.InitSysConfig()
	runDebugHttpServer()
	return runXdsServer()
}

var (
	currentVersion = 1
	snapshotCache  cache.SnapshotCache
	nodeID         = "test1" // 测试用的节点ID
)

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
	snapshotCache = cache.NewSnapshotCache(false, cache.IDHash{}, llog)

	// envoy配置的缓存快照
	//snapshot := utils.GenerateSnapshot(strconv.Itoa(currentVersion))
	snapshot := utils.NewEmptySnapshot(strconv.Itoa(currentVersion))
	if err := snapshot.Consistent(); err != nil {
		llog.Errorf("snapshot inconsistency: %+v\n%+v", snapshot, err)
		os.Exit(1)
	}

	// Add the snapshot to the cache
	if err := snapshotCache.SetSnapshot(context.Background(), nodeID, snapshot); err != nil {
		os.Exit(1)
	}

	// 请求回调
	cb := &utils.Callbacks{Debug: llog.Debug}
	// 官方提供的控制面server
	srv := server.NewServer(context.Background(), snapshotCache, cb)
	// 注册
	routeservice.RegisterRouteDiscoveryServiceServer(grpcServer, srv)
	listenerservice.RegisterListenerDiscoveryServiceServer(grpcServer, srv)
	clusterservice.RegisterClusterDiscoveryServiceServer(grpcServer, srv)

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

// ApplyIngress 新增/更新envoy配置
func ApplyIngress(ing *v1.Ingress) {
	currentVersion++
	newSnapshot := utils.ApplySnapshot(strconv.Itoa(currentVersion), ing)
	if err := snapshotCache.SetSnapshot(context.Background(), nodeID, newSnapshot); err != nil {
		log.Println("apply ingress error:", err)
	}
}

// RemoveIngress 移除该ingress的envoy配置
func RemoveIngress(ing *v1.Ingress) {
	currentVersion++
	newSnapshot := utils.RemoveSnapshot(strconv.Itoa(currentVersion), ing)
	if err := snapshotCache.SetSnapshot(context.Background(), nodeID, newSnapshot); err != nil {
		log.Println("remove ingress error:", err)
	}
}

// 启动http服务，用于重载配置
func runDebugHttpServer() {
	go func() {
		r := gin.New()
		r.GET("/reload", func(c *gin.Context) {
			currentVersion++
			ss := utils.GenerateTestSnapshot(strconv.Itoa(currentVersion))
			err := snapshotCache.SetSnapshot(c, nodeID, ss)
			if err != nil {
				c.String(400, err.Error())
			} else {
				c.String(200, "success")
			}
		})
		fmt.Println("启动Debug Server，端口是18000")
		if err := r.Run(":18000"); err != nil {
			log.Println("debug server 启动失败")
		}
	}()
}
