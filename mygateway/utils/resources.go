package utils

import (
	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"istio-envoy/mygateway/utils/helpers"
	v1 "k8s.io/api/networking/v1"
	"log"
	"sigs.k8s.io/yaml"
	"sync"
)

// 每个ingress资源对应的快照 key: ingress.Uid value: *cache.Snapshot
var ingressSnapshots sync.Map

type snapshot struct {
	listeners, routeConfigs, clusters []types.Resource
}

func newIngressSnapshot(lis *listener.Listener, rc *route.RouteConfiguration, cls []*cluster.Cluster) *snapshot {
	snap := &snapshot{}
	snap.listeners = append(snap.listeners, lis)
	snap.routeConfigs = append(snap.routeConfigs, rc)
	for _, c := range cls {
		snap.clusters = append(snap.clusters, c)
	}
	return snap
}

// 生成ingress资源的快照
func generateIngressSnapshot(ingress *v1.Ingress) *snapshot {
	// 渲染listener
	listeners := &listener.Listener{}
	if err := NewTplGenerator[*listener.Listener]().GetOutput(ingress, "listener", listeners); err != nil {
		log.Println(err)
	}

	// 渲染routeConfig
	routeConfigs := &route.RouteConfiguration{}
	if err := NewTplGenerator[*route.RouteConfiguration]().GetOutput(ingress, "route", routeConfigs); err != nil {
		log.Fatalln(err)
	}

	// 渲染clusters
	cls, err := NewTplGenerator[*cluster.Cluster]().
		GetOutputs(ingress, "clusters", func() *cluster.Cluster {
			return &cluster.Cluster{}
		})
	if err != nil {
		log.Fatalln(err)
	}

	return newIngressSnapshot(listeners, routeConfigs, cls)
}

// 获取所有ingress的快照
func getAllSnapshot() *snapshot {
	snapshots := &snapshot{}
	ingressSnapshots.Range(func(key, value any) bool {
		if snap, ok := value.(*snapshot); ok {
			snapshots.listeners = append(snapshots.listeners, snap.listeners...)
			snapshots.routeConfigs = append(snapshots.routeConfigs, snap.routeConfigs...)
			snapshots.clusters = append(snapshots.clusters, snap.clusters...)
		}

		return true
	})

	return snapshots
}

// ApplySnapshot 新增或更新ingress快照
func ApplySnapshot(version string, ingress *v1.Ingress) *cache.Snapshot {
	// 设置该ingress快照缓存
	ingressSnapshots.Store(ingress.UID, generateIngressSnapshot(ingress))

	// 获取所有快照
	allSnaps := getAllSnapshot()

	snapCache, _ := cache.NewSnapshot(version,
		map[resource.Type][]types.Resource{
			resource.ClusterType:  allSnaps.clusters,
			resource.RouteType:    allSnaps.routeConfigs,
			resource.ListenerType: allSnaps.listeners,
		},
	)
	return snapCache
}

// RemoveSnapshot 移除ingress快照
func RemoveSnapshot(version string, ingress *v1.Ingress) *cache.Snapshot {
	// 删除该ingress快照缓存
	ingressSnapshots.Delete(ingress.UID)

	// 获取所有快照
	allSnaps := getAllSnapshot()

	snap, _ := cache.NewSnapshot(version,
		map[resource.Type][]types.Resource{
			resource.ClusterType:  allSnaps.clusters,
			resource.RouteType:    allSnaps.routeConfigs,
			resource.ListenerType: allSnaps.listeners,
		},
	)
	return snap
}

// NewEmptySnapshot 创建一个空的快照对象
func NewEmptySnapshot(version string) *cache.Snapshot {
	snap, _ := cache.NewSnapshot(version,
		map[resource.Type][]types.Resource{
			resource.ClusterType:  {},
			resource.RouteType:    {},
			resource.ListenerType: {},
		},
	)
	return snap
}

const TestIngress = "mygateway/test/ingress.yaml"

// GenerateTestSnapshot 解析测试yaml生成快照
func GenerateTestSnapshot(version string) *cache.Snapshot {
	// 把ingress yaml反序列化为对象
	ingBytes := helpers.MustLoadFile(TestIngress)
	ingress := &v1.Ingress{}
	if err := yaml.Unmarshal(ingBytes, ingress); err != nil {
		log.Fatalln(err)
	}

	// 渲染listener
	lis := &listener.Listener{}
	err := NewTplGenerator[*listener.Listener]().
		GetOutput(ingress, "listener", lis)
	if err != nil {
		log.Fatalln(err)
	}

	// 渲染routeConfig
	routeConfig := &route.RouteConfiguration{}
	err = NewTplGenerator[*route.RouteConfiguration]().
		GetOutput(ingress, "route", routeConfig)
	if err != nil {
		log.Fatalln(err)
	}

	// 渲染clusters
	var resList []types.Resource
	clusters, err := NewTplGenerator[*cluster.Cluster]().
		GetOutputs(ingress, "clusters", func() *cluster.Cluster {
			return &cluster.Cluster{}
		})
	if err != nil {
		log.Fatalln(err)
	}
	for _, c := range clusters {
		resList = append(resList, c)
	}

	// 构建snapshot
	snap, _ := cache.NewSnapshot(version,
		map[resource.Type][]types.Resource{
			resource.ClusterType:  resList,
			resource.RouteType:    {routeConfig},
			resource.ListenerType: {lis},
		},
	)
	return snap
}
